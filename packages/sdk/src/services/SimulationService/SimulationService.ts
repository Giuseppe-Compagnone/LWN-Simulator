import {
  SimulationConfig,
  SimulationEvent,
  SimulationSnapshot,
} from "@lwn-simulator/contracts";
import { ApiCaller, BaseService } from "../../models";

export class SimulationService extends BaseService {
  private readonly baseUrl: string;

  constructor(baseUrl: string) {
    super("simulation", baseUrl);
    this.baseUrl = baseUrl;
  }

  public start = async (config: SimulationConfig): Promise<SimulationSnapshot> =>
    this.apiCaller.post<SimulationSnapshot>("/start", config);

  public pause = async (): Promise<SimulationSnapshot> =>
    this.apiCaller.post<SimulationSnapshot>("/pause");

  public resume = async (): Promise<SimulationSnapshot> =>
    this.apiCaller.post<SimulationSnapshot>("/resume");

  public stop = async (): Promise<SimulationSnapshot> =>
    this.apiCaller.post<SimulationSnapshot>("/stop");

  public getSnapshot = async (): Promise<SimulationSnapshot> =>
    this.apiCaller.get<SimulationSnapshot>("/snapshot");

  public getWebSocketUrl = (): string => {
    const url = new URL(ApiCaller.joinUrl(this.baseUrl, "/simulation/ws"));
    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    return url.toString();
  };

  public static appendEvent(
    events: Array<SimulationEvent>,
    event: SimulationEvent,
    limit = 100,
  ): Array<SimulationEvent> {
    return [...events, event].slice(-limit);
  }
}
