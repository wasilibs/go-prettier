FROM oven/bun:1 AS app

RUN apt-get update && apt-get install -y build-essential binaryen cmake curl

WORKDIR /quickjs

RUN curl -L https://github.com/quickjs-ng/quickjs/archive/76b3b124000886de9422a82e4f9524ec30dc36f0.tar.gz | tar -xz --strip-components=1 -C /quickjs
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build --target qjsc

WORKDIR /app

COPY buildtools/wasm/package.json buildtools/wasm/bun.lock ./
RUN bun install --frozen-lockfile --production
COPY buildtools/wasm/*.ts ./
COPY buildtools/wasm/go/ ./go
COPY buildtools/wasm/sh/ ./sh
COPY buildtools/wasm/tsconfig.json ./
RUN bun run build

FROM ghcr.io/webassembly/wasi-sdk:sha-d90d7de

RUN apt-get update && apt-get install -y binaryen curl

WORKDIR /quickjs

RUN curl -L https://github.com/quickjs-ng/quickjs/archive/76b3b124000886de9422a82e4f9524ec30dc36f0.tar.gz | tar -xz --strip-components=1 -C /quickjs
ENV CMAKE_TOOLCHAIN_FILE=/opt/wasi-sdk/share/cmake/wasi-sdk-p1.cmake
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release -DQJS_BUILD_LIBC=ON && cmake --build build --target qjs

COPY --from=app /app/dist/prettier.js .
COPY --from=app /quickjs/build/qjsc .

RUN ./qjsc -o prettier.c -e prettier.js
# Make sure LLVM stack size matches qjsc default.
RUN $CC $CFLAGS -O3 -o prettier-noopt.wasm prettier.c -L/quickjs/build -lqjs -I/quickjs -Wl,-lwasi-emulated-signal -Wl,-z,stack-size=1048576

RUN wasm-opt -o prettier.wasm --low-memory-unused --flatten --rereloop --converge -O3 prettier-noopt.wasm

CMD [ "cp", "prettier.wasm", "/out"]
