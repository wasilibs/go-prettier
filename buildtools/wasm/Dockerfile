FROM oven/bun:1-alpine AS buildjs

WORKDIR /app

COPY buildtools/wasm/package.json buildtools/wasm/bun.lock ./
RUN bun install --frozen-lockfile --production
COPY buildtools/wasm/*.ts ./
COPY buildtools/wasm/go/ ./go
COPY buildtools/wasm/sh/ ./sh
COPY buildtools/wasm/tsconfig.json ./
RUN bun run build

FROM ghcr.io/webassembly/wasi-sdk:wasi-sdk-25

RUN apt-get update && apt-get install -y binaryen curl patch xz-utils

WORKDIR /quickjs
RUN curl -L https://github.com/bellard/quickjs/archive/f25e5d4094a11cf098670417e8a16ffb7cbadda0.tar.gz | tar -xz --strip-components=1 -C /quickjs
RUN make
RUN make install
RUN make clean
COPY buildtools/wasm/patch-quickjs.txt ./
RUN patch -p1 < patch-quickjs.txt
ENV CUSTOM_CFLAGS $CFLAGS -O3 -DEMSCRIPTEN -D_WASI_EMULATED_SIGNAL
RUN make libquickjs.a
RUN $RANLIB libquickjs.a

WORKDIR /app
COPY --from=buildjs /app/dist/prettier.js ./
RUN qjsc -e prettier.js -o prettier.c
RUN $CC $CFLAGS -O3 -o prettier-noopt.wasm prettier.c -lquickjs -I/usr/local/include/quickjs -L/quickjs -Wl,--global-base=1024 -Wl,-lwasi-emulated-signal
RUN wasm-opt -o prettier.wasm --low-memory-unused --flatten --rereloop --converge -O3 prettier-noopt.wasm

CMD [ "cp", "prettier.wasm", "/out"]
