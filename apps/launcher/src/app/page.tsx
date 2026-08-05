"use client";

import {
  Card,
  Logo,
  LogoLayout,
  LogoSize,
  Form,
  FormValue,
  CardType,
  CardLayout,
  radioField,
  textField,
  FormField,
  Spinner,
  SpinnerSize,
  NotificationHandler,
} from "@lwn-simulator/ui-components";
import { useState } from "react";

export default function Home() {
  // States
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [loadingText, setLoadingText] = useState<string>("");

  return (
    <div className={`home-page page`}>
      {!isLoading ? (
        <>
          <Logo size={LogoSize.Lg} layout={LogoLayout.SecondaryBackground} />
          <h2 className="title">Choose a Connection</h2>
          <h3 className="sub-title">
            Select a local backend or connect to a remote server
          </h3>{" "}
          <Card
            type={CardType.Transparent}
            layout={CardLayout.Padded}
            className="connection-card"
          >
            <Form
              fields={[
                radioField({
                  name: "env",
                  label: "Server Environment",
                  value: null,
                  error: null,
                  options: [
                    {
                      value: "local",
                      displayed: <>Local Backend</>,
                    },
                    {
                      value: "remote",
                      displayed: <>Remote Server</>,
                    },
                  ],
                  required: true,
                }),
                textField({
                  name: "url",
                  label: "Remote Url",
                  value: "",
                  placeholder: "https://example.com:8080",
                  error: null,
                  required: true,
                  display: (fieldsState: Record<string, FormField>) =>
                    fieldsState["env"].value === "remote",
                  validations: [
                    {
                      rule: /^(?:(?:https?):\/\/)?(?:localhost|(?:\d{1,3}\.){3}\d{1,3}|(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,})(?::\d{1,5})?$/,
                      error: "Invalid url",
                    },
                  ],
                }),
              ]}
              onSubmit={async (
                values: Record<string, FormValue>,
              ): Promise<void> => {
                switch (values["env"]) {
                  case "local":
                    console.log("[Launcher] Running local backend...");
                    setLoadingText("Running local backend");
                    setIsLoading(true);
                    window.electron.connectLocal();
                    break;
                  case "remote":
                    setLoadingText(`Connecting to: ${values["url"]}`);
                    setIsLoading(true);
                    console.log(`[Launcher] Connecting to ${values["url"]}...`);
                    const result = await window.electron.connectRemote(
                      values["url"] as string,
                    );
                    console.log("\n\nRESULT", result);
                    if (!result.success) {
                      setLoadingText("");
                      setIsLoading(false);
                      NotificationHandler.instance.error(
                        result.message || "Error",
                      );
                    }
                    break;
                  default:
                    break;
                }
              }}
            />
          </Card>
        </>
      ) : (
        <>
          <Spinner size={SpinnerSize.Lg} />
          <p className="loading-text">{loadingText}</p>
        </>
      )}
    </div>
  );
}
