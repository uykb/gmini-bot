
import './wasm_exec.js';
import goWasm from './main.wasm';

const go = new Go();

// Instantiate the Wasm module once
const instantiateWasm = WebAssembly.instantiate(goWasm, go.importObject).then(instance => {
  go.run(instance);
  console.log("Go Wasm module instantiated and running.");
  return instance;
}).catch(err => {
  console.error("Wasm instantiation failed:", err);
  throw err; // Throw error to prevent worker from running with a faulty module
});

export default {
  /**
   * @param {ScheduledEvent} event
   * @param {Env} env
   * @param {ExecutionContext} ctx
   */
  async scheduled(event, env, ctx) {
    console.log(`Cron triggered: ${event.cron}`);

    try {
      // Wait for the Wasm module to be ready
      await instantiateWasm;

      // Pass environment variables and bindings to the Go environment.
      // This makes them accessible via `js.Global().Get()` in Go.
      go.env = {
        ...env, // Pass all wrangler secrets and vars
      };

      // Also attach the KV namespace binding to the global scope for Go to access.
      // Note: Cloudflare bindings are not serializable, so they can't be in `go.env`.
      if (env.SIGNAL_CACHE) {
        globalThis.SIGNAL_CACHE = env.SIGNAL_CACHE;
      }

      // Check if the 'run' function is exported from Go
      if (typeof run === "function") {
        console.log("Calling the 'run' function exported from Go...");
        run();
      } else {
        console.error("'run' function not found. Was it exported from Go correctly?");
      }
      
    } catch (error) {
      console.error("Error during scheduled execution:", error);
    }
  },
};
