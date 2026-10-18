import { describe, expect, it } from "vitest";
import { isValidJavaScript } from "./payloadCodec";

describe("isValidJavaScript", () => {
  it("accepts a valid payload decoder", () => {
    expect(
      isValidJavaScript(
        "function decodeUplink(input) { return { data: input.bytes }; }",
      ),
    ).toBe(true);
  });

  it("rejects an invalid payload decoder", () => {
    expect(
      isValidJavaScript("function decodeUplink(input) { return { data: }; }"),
    ).toBe(false);
  });
});
