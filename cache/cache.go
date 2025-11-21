
// Package cache provides functions to interact with a Cloudflare KV namespace
// for caching signals to avoid duplicate notifications.
package cache

import (
	"fmt"
	"syscall/js"
)

// await wraps a JavaScript Promise in a Go channel, allowing to wait for its resolution.
func await(promise js.Value) (js.Value, error) {
	resCh := make(chan js.Value, 1)
	errCh := make(chan error, 1)

	onSuccess := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			resCh <- args[0]
		} else {
			resCh <- js.Undefined()
		}
		return nil
	})
	defer onSuccess.Release()

	onError := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			errCh <- fmt.Errorf("promise rejected: %s", args[0].String())
		} else {
			errCh <- fmt.Errorf("promise rejected with no error message")
		}
		return nil
	})
	defer onError.Release()

	promise.Call("then", onSuccess, onError)

	select {
	case res := <-resCh:
		return res, nil
	case err := <-errCh:
		return js.Undefined(), err
	}
}

// GetKVNamespace retrieves the KV namespace binding from the global JavaScript scope.
func GetKVNamespace(bindingName string) (js.Value, error) {
	kv := js.Global().Get(bindingName)
	if kv.IsUndefined() || kv.IsNull() {
		return js.Value{}, fmt.Errorf("KV namespace binding '%s' not found in global scope", bindingName)
	}
	return kv, nil
}

// KeyExists checks if a key exists in the KV namespace.
func KeyExists(kv js.Value, key string) (bool, error) {
	promise := kv.Call("get", key)
	value, err := await(promise)
	if err != nil {
		return false, fmt.Errorf("failed to get key '%s' from KV: %w", key, err)
	}
	// .get() returns null if the key doesn't exist, which IsNull() correctly checks.
	return !value.IsNull(), nil
}

// SetKey sets a key in the KV namespace with a specified TTL in seconds.
func SetKey(kv js.Value, key string, ttlSeconds int) error {
	options := js.Global().Get("Object").New()
	options.Set("expirationTtl", ttlSeconds)

	promise := kv.Call("put", key, "1", options)
	_, err := await(promise)
	if err != nil {
		return fmt.Errorf("failed to set key '%s' in KV: %w", key, err)
	}
	return nil
}
