import fs from "fs";

const input = "generated/openapi.ts";
const output = "generated/models.ts";

const content = fs.readFileSync(input, "utf8");

const schemasSection = content.match(
  /export interface components \{[\s\S]*?schemas: \{([\s\S]*?)\n    \};/,
);

if (!schemasSection) {
  throw new Error("Schemas section not found");
}

const schemas = [
  ...schemasSection[1].matchAll(/^\s{8}(\w+): \{([\s\S]*?)^\s{8}\};/gm),
];

let result = "";

for (const schema of schemas) {
  const name = schema[1].charAt(0).toUpperCase() + schema[1].slice(1);
  const body = schema[2];

  result += `
export interface ${name} {${body}}
`;
}

fs.writeFileSync(output, result);

console.log(`Generated ${output}`);
