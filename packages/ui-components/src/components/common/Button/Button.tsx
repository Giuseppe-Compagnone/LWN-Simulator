"use client";

import { useState } from "react";
import {
  ButtonFont,
  ButtonLayout,
  ButtonProps,
  ButtonType,
} from "./Button.types";
import cn from "classnames";
import Spinner, { SpinnerSize, SpinnerType } from "../Spinner";

const Button = (props: ButtonProps) => {
  const type = props.type || ButtonType.Primary;
  const font = props.font || ButtonFont.Primary;
  const layout = props.layout || ButtonLayout.Default;

  //States
  const [isLoading, setIsLoading] = useState<boolean>(false);

  return (
    <button
      onClick={async (e) => {
        setIsLoading(true);
        await props.onClick?.(e);
        setIsLoading(false);
      }}
      className={cn(
        "button",
        props.className,
        type,
        layout,
        props.disabled && "disabled",
        isLoading && "loading",
      )}
      style={{ fontFamily: `var(${font})` }}
    >
      {props.value}
      {isLoading && (
        <div className="loading-wrapper">
          <Spinner
            size={SpinnerSize.Sm}
            type={
              type === ButtonType.Primary
                ? SpinnerType.Secondary
                : SpinnerType.Primary
            }
          />
        </div>
      )}
    </button>
  );
};

export default Button;
