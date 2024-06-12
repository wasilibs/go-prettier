import { clearTimeout, setTimeout } from "os";

(globalThis as any).clearTimeout = clearTimeout;
(globalThis as any).setTimeout = setTimeout;
