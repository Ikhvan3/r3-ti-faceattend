import { existsSync } from "node:fs";
import { registerHooks } from "node:module";
import { fileURLToPath, pathToFileURL } from "node:url";

registerHooks({
  resolve(specifier, context, nextResolve) {
    if (
      specifier.startsWith(".") &&
      !specifier.endsWith(".ts") &&
      !specifier.endsWith(".tsx") &&
      !specifier.endsWith(".mjs") &&
      !specifier.endsWith(".js")
    ) {
      const parentPath = fileURLToPath(context.parentURL);
      const candidate = new URL(specifier + ".ts", pathToFileURL(parentPath));
      if (existsSync(candidate)) {
        return nextResolve(candidate.href, context);
      }
    }
    return nextResolve(specifier, context);
  },
});
