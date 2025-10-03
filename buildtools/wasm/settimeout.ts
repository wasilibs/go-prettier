import { clearTimeout, setTimeout } from "qjs:os";

(globalThis as any).clearTimeout = clearTimeout;
(globalThis as any).setTimeout = setTimeout;
