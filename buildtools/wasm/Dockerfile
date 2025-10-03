FROM ghcr.io/webassembly/wasi-sdk:wasi-sdk-27 AS qjs

RUN apt-get update && apt-get install -y curl

WORKDIR /quickjs

RUN curl -L https://github.com/quickjs-ng/quickjs/archive/76b3b124000886de9422a82e4f9524ec30dc36f0.tar.gz | tar -xz --strip-components=1 -C /quickjs
ENV CMAKE_TOOLCHAIN_FILE=/opt/wasi-sdk/share/cmake/wasi-sdk.cmake
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build --target qjs_exe

FROM oven/bun:1

RUN apt-get update && apt-get install -y build-essential binaryen cmake curl

WORKDIR /quickjs

RUN curl -L https://github.com/quickjs-ng/quickjs/archive/76b3b124000886de9422a82e4f9524ec30dc36f0.tar.gz | tar -xz --strip-components=1 -C /quickjs
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build --target qjs_exe

WORKDIR /app

COPY buildtools/wasm/package.json buildtools/wasm/bun.lock ./
RUN bun install --frozen-lockfile --production
COPY buildtools/wasm/*.ts ./
COPY buildtools/wasm/go/ ./go
COPY buildtools/wasm/sh/ ./sh
COPY buildtools/wasm/tsconfig.json ./
RUN bun run build

WORKDIR /quickjs

COPY --from=qjs /quickjs/build/qjs ./qjs-wasm
RUN build/qjs -c /app/dist/prettier.js -o prettier.wasm --exe ./qjs-wasm

#RUN wasm-opt -o prettier.wasm --low-memory-unused --flatten --rereloop --converge -O3 prettier-noopt.wasm

CMD [ "cp", "prettier.wasm", "/out"]
