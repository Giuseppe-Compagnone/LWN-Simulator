import { MouseEvent } from "react";

export enum ButtonType {
  Primary = "primary",
  Secondary = "secondary",
  Outlined = "outlined",
}

export enum ButtonFont {
  Primary = "--primary-font",
  Secondary = "--secondary-font",
}

export interface ButtonProps {
  value: string;
  type?: ButtonType;
  font?: ButtonFont;
  block?: boolean;
  onClick?: (e?: MouseEvent) => void | Promise<void>;
}
