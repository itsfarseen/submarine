import * as definitions from "@polkadot/types/interfaces/definitions"
import * as fs from "node:fs"
import * as path from "node:path"

import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const output_path = path.join(__dirname, "..", "schema.json")

console.log("Available modules:")
let modules = {}
for (let key of Object.keys(definitions)) {
	console.log("-", key)
	modules[key] = definitions[key].types
}

let json = JSON.stringify(modules, null, 2)
fs.writeFileSync(output_path, json)
console.log("Saved to", output_path)
