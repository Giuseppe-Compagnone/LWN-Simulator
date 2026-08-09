import axios, { AxiosRequestConfig } from "axios";

export class ApiCaller {
  constructor(private readonly baseUrl: string) {}

  private async request<T>(config: AxiosRequestConfig): Promise<T> {
    const response = await axios<T>({
      ...config,
      baseURL: this.baseUrl,
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
