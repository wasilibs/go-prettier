import type { AstPath, Doc, Parser, Plugin, Printer, SupportLanguage } from "prettier";
import { in as stdin, out as stdout } from "std";

type StringNode = {
  body: string;
  start: number;
  end: number;
};

const goParser: Parser = {
  astFormat: "go",
  locStart: (node: StringNode) => node.start,
  locEnd: (node: StringNode) => node.end,
  parse(text: string): StringNode {
    return {
      body: text,
      start: 0,
      end: text.length,
    };
  },
};

const languages: SupportLanguage[] = [
  {
    name: "Go",
    parsers: ["go"],
    extensions: [".go"],
  },
];

const parsers: Record<string, Parser> = {
  go: goParser,
};

const goPrinter: Printer = {
  print(path: AstPath): Doc {
    const node: StringNode = path.node;
    const msg = {
      name: "gofmt-request",
      body: node.body,
    };
    const msgStr = JSON.stringify(msg);
    stdout.printf("%s\n", msgStr);
    stdout.flush();
    const responseStr = stdin.getline();
    const response = JSON.parse(responseStr);
    if (response.name !== "gofmt-response") {
      throw new Error(`Expected response name to be "gofmt-result", got "${response.name}"`);
    }
    return response.body;
  },
};

const printers: Record<string, Printer> = {
  go: goPrinter,
};

const plugin: Plugin<StringNode> = {
  languages,
  parsers,
  printers,
}

export default plugin;
