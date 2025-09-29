//go:build integration

package keyservice_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/illmade-knight/go-key-service/test"
	"github.com/illmade-knight/go-microservice-base/pkg/middleware"
	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Setup Helpers ---

// newJWKSTestServer creates a mock HTTP server that serves a JWKS containing
// the public part of the provided private key.
func newJWKSTestServer(t *testing.T, privateKey *rsa.PrivateKey) *httptest.Server {
	t.Helper()

	publicKey, err := jwk.FromRaw(privateKey.Public())
	require.NoError(t, err)
	_ = publicKey.Set(jwk.KeyIDKey, "test-key-id")
	_ = publicKey.Set(jwk.AlgorithmKey, jwa.RS256)

	keySet := jwk.NewSet()
	_ = keySet.AddKey(publicKey)

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(keySet)
		require.NoError(t, err)
	})
	return httptest.NewServer(mux)
}

// createTestToken generates a valid RS256 JWT for testing.
func createTestToken(t *testing.T, privateKey *rsa.PrivateKey, userID string) string {
	t.Helper()
	token, err := jwt.NewBuilder().
		Subject(userID).
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(time.Hour)).
		Build()
	require.NoError(t, err)

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, privateKey))
	require.NoError(t, err)
	return string(signed)
}

// --- Main Test ---

func TestServiceIntegration(t *testing.T) {
	// --- 1. Setup ---
	// Generate a real RSA key pair for this test run.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a mock JWKS server that will serve the public key.
	jwksServer := newJWKSTestServer(t, privateKey)
	t.Cleanup(jwksServer.Close)

	// Create the real authentication middleware, pointing it to our mock server.
	authMiddleware, err := middleware.NewJWKSAuthMiddleware(jwksServer.URL)
	require.NoError(t, err)

	// Create the key service test server, injecting our real middleware.
	keyServiceServer := test.NewTestServer(authMiddleware)
	t.Cleanup(keyServiceServer.Close)

	// --- 2. Test Cases ---

	t.Run("StoreKey - Success with valid token", func(t *testing.T) {
		// Arrange
		testURN, _ := urn.New(urn.SecureMessaging, "user", "user-123")
		token := createTestToken(t, privateKey, "user-123")
		req, _ := http.NewRequest(http.MethodPost, keyServiceServer.URL+"/keys/"+testURN.String(), bytes.NewBufferString("my-public-key"))
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		// Assert
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("StoreKey - Failure when JWT subject does not match URN", func(t *testing.T) {
		// Arrange
		testURN, _ := urn.New(urn.SecureMessaging, "user", "user-123")
		// Token is for a DIFFERENT user
		token := createTestToken(t, privateKey, "another-user-456")
		req, _ := http.NewRequest(http.MethodPost, keyServiceServer.URL+"/keys/"+testURN.String(), bytes.NewBufferString("my-public-key"))
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		// Assert
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("GetKey - Success (public endpoint)", func(t *testing.T) {
		// This endpoint requires no token. First, store a key to retrieve.
		storeURN, _ := urn.New(urn.SecureMessaging, "user", "user-to-get")
		storeToken := createTestToken(t, privateKey, "user-to-get")
		storeReq, _ := http.NewRequest(http.MethodPost, keyServiceServer.URL+"/keys/"+storeURN.String(), bytes.NewBufferString("the-key-to-find"))
		storeReq.Header.Set("Authorization", "Bearer "+storeToken)
		_, _ = http.DefaultClient.Do(storeReq)

		// Now, try to get it without a token.
		getReq, _ := http.NewRequest(http.MethodGet, keyServiceServer.URL+"/keys/"+storeURN.String(), nil)
		resp, err := http.DefaultClient.Do(getReq)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		// Assert
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "the-key-to-find", string(body))
	})
}
