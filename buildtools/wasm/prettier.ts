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
import { exit, err as stderr, in as stdin, out as stdout } from "std";

async function run() {
  const config = JSON.parse(scriptArgs[1]);

  const content = stdin.readAsString();

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
        pluginHtml,
        pluginGraphQl,
        pluginMarkdown,
        pluginMeriyah,
        pluginPostcss,
        pluginTypescript,
        pluginYaml,
      ],
    });
  } catch (e: any) {
    if (e.name === "UndefinedParserError") {
      exit(10);
    }
    stderr.printf("%s", e.message);
    exit(1);
  }

  stdout.puts(response);
}

await run();
