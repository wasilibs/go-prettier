import "./global.js";
import "./settimeout.js";
import "./textcoding.js";

import { format } from "prettier";
import pluginAcorn from "prettier/plugins/acorn.js";
import pluginAngular from "prettier/plugins/angular.js";
import pluginBabel from "prettier/plugins/babel.js";
import pluginEsTree from "prettier/plugins/estree.js";
import pluginGlimmer from "prettier/plugins/glimmer.js";
import pluginGraphQl from "prettier/plugins/graphql.js";
import pluginHtml from "prettier/plugins/html.js";
import pluginMarkdown from "prettier/plugins/markdown.js";
import pluginMeriyah from "prettier/plugins/meriyah.js";
import pluginPostcss from "prettier/plugins/postcss.js";
import pluginTypescript from "prettier/plugins/typescript.js";
import pluginYaml from "prettier/plugins/yaml.js";

import pluginGo from "./go/index.js";
import pluginSh from "./sh/index.js";

import { exit, err as stderr, in as stdin, out as stdout } from "qjs:std";

async function run() {
  stdout.printf("running\n");
  const config = JSON.parse(scriptArgs[1]);

  stdout.printf("getting input\n");
  const inputStr = stdin.getline();
  stdout.printf("got input: %s\n", inputStr);
  const inputMsg = JSON.parse(inputStr);
  const content = inputMsg.body;
  stdout.printf("got content: %s\n", content);

  let response: string;

  try {
    response = await format(content, {
      ...config,
      plugins: [
        pluginAcorn,
        pluginAngular,
        pluginBabel,
        pluginEsTree,
        pluginGlimmer,
        pluginGo,
        pluginHtml,
        pluginGraphQl,
        pluginMarkdown,
        pluginMeriyah,
        pluginPostcss,
        pluginSh,
        pluginTypescript,
        pluginYaml,
      ],
    });
  } catch (e: any) {
    stdout.printf("caught error: %s\n", e);
    if (e.name === "UndefinedParserError") {
      exit(10);
    }
    stdout.printf("%s\n", e.message);
    exit(1);
  }
  stdout.printf("formatted content: %s\n", response);

  const outputMsg = {
    name: "result",
    body: response,
  };
  const outputStr = JSON.stringify(outputMsg);
  stderr.printf("%s\n", outputStr);
  stdout.printf("%s\n", outputStr);
}

await run();
stdout.printf("done run\n");

stderr.flush();
stdout.flush();
