import type { SimulationEvent } from "@lwn-simulator/contracts";
import { describe, expect, it } from "vitest";
import { SimulationService } from "./SimulationService";

const event = (id: string, sequence: number): SimulationEvent =>
  ({ id, sequence, message: id } as SimulationEvent);

describe("SimulationService event collection", () => {
  it("merges by event id, orders by sequence, and keeps the newest limit", () => {
    const current = [event("one", 1), event("two", 2)];
    const incoming = [event("two", 3), event("three", 4)];

    expect(SimulationService.mergeEvents(current, incoming, 2)).toEqual([
      event("two", 3),
      event("three", 4),
    ]);
  });

  it("appends an event without duplicating an existing id", () => {
    expect(
      SimulationService.appendEvent([event("one", 1)], event("one", 2)),
    ).toEqual([event("one", 2)]);
  });
});
