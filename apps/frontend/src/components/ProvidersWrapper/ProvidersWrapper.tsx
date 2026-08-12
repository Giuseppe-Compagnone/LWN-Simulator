"use client";

import { ThemeServiceProvider } from "@lwn-simulator/ui-components";
import { ProvidersWrapperProps } from "./ProvidersWrapper.types";
import {
  AppInfoServiceProvider,
  DeviceServiceProvider,
} from "@lwn-simulator/sdk";
import Navbar from "../Navbar";
import Sidebar from "../Sidebar";
import Footer from "../Footer";
import { useEffect, useState } from "react";

const ProvidersWrapper = (props: ProvidersWrapperProps) => {
  // States
  const [origin, setOrigin] = useState("http://localhost:8080/api");

  // Effects
  useEffect(() => {
    setOrigin(
      `${process.env.NODE_ENV === "development" ? "http://localhost:8080" : window.location.origin}/api`,
    );
  }, []);
  return (
    <ThemeServiceProvider>
      <AppInfoServiceProvider baseUrl={origin}>
        <DeviceServiceProvider>
          <Navbar />
          <Sidebar>
            {props.children}
            <Footer />
          </Sidebar>
        </DeviceServiceProvider>
      </AppInfoServiceProvider>
    </ThemeServiceProvider>
  );
};

export default ProvidersWrapper;
