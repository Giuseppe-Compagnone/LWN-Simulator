"use client";

import { Spinner, SpinnerSize } from "@lwn-simulator/ui-components";
import { PageLoaderProps } from "./PageLoader.types";
import { useEffect } from "react";
import { redirect } from "next/navigation";

const PageLoader = (props: PageLoaderProps) => {
  useEffect(() => {
    redirect(props.href);
  }, [props.href]);

  return (
    <div className="page page-loader">
      <Spinner size={SpinnerSize.Lg} />
    </div>
  );
};

export default PageLoader;
