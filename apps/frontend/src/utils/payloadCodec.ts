import { parse } from "acorn";

export const isValidJavaScript = (value: string): boolean => {
  try {
    parse(value, { ecmaVersion: "latest", sourceType: "script" });
    return true;
  } catch {
    return false;
  }
};
