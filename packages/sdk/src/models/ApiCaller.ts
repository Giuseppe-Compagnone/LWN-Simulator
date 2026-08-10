import axios, { AxiosRequestConfig } from "axios";

export class ApiCaller {
  constructor(private readonly serviceName: string) {
    this.serviceName = serviceName;
  }

  public static baseUrl: string | null = null;

  public static joinUrl(base: string, path: string) {
    return new URL(
      path.replace(/^\/+/, ""),
      `${base.replace(/\/+$/, "")}/`,
    ).toString();
  }

  private async request<T>(config: AxiosRequestConfig): Promise<T> {
    if (!ApiCaller.baseUrl) throw new Error("BaseUrl is undefined");

    const response = await axios<T>({
      ...config,
      baseURL: ApiCaller.joinUrl(ApiCaller.baseUrl, this.serviceName),
    });

    return response.data;
  }

  public get<T>(path: string, config?: AxiosRequestConfig): Promise<T> {
    return this.request<T>({
      method: "GET",
      url: path,
      ...config,
    });
  }

  public post<T>(
    path: string,
    data?: unknown,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return this.request<T>({
      method: "POST",
      url: path,
      data,
      ...config,
    });
  }

  public put<T>(
    path: string,
    data?: unknown,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return this.request<T>({
      method: "PUT",
      url: path,
      data,
      ...config,
    });
  }

  public delete<T>(path: string, config?: AxiosRequestConfig): Promise<T> {
    return this.request<T>({
      method: "DELETE",
      url: path,
      ...config,
    });
  }
}
