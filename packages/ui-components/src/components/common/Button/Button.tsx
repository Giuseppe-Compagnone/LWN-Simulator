import { useState } from "react";
import { ButtonFont, ButtonProps, ButtonType } from "./Button.types";
import cn from "classnames";
import Spinner, { SpinnerSize, SpinnerType } from "../Spinner";

const Button = (props: ButtonProps) => {
  const type = props.type || ButtonType.Primary;
  const font = props.font || ButtonFont.Primary;

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
        type,
        props.block && "block",
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
