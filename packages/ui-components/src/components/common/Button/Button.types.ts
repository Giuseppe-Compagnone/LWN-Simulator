import { MouseEvent } from "react";

/**
 * Defines the visual style variant of a button.
 */
export enum ButtonType {
  /** Main action button style. */
  Primary = "primary",

  /** Secondary action button style. */
  Secondary = "secondary",

  /** Button style with transparent background and visible border. */
  Outlined = "outlined",
}

/**
 * Defines the font family used by the button text.
 */
export enum ButtonFont {
  /** Uses the primary font family. */
  Primary = "--primary-font",

  /** Uses the secondary font family. */
  Secondary = "--secondary-font",
}

/**
 * Properties for configuring a button component.
 */
export interface ButtonProps {
  /**
   * Text content displayed inside the button.
   */
  value: string;

  /**
   * Additional CSS class names applied to the button element.
   */
  className?: string;

  /**
   * Visual style variant of the button.
   *
   * @default ButtonType.Primary
   */
  type?: ButtonType;

  /**
   * Font family used for the button text.
   *
   * @default ButtonFont.Primary
   */
  font?: ButtonFont;

  /**
   * Expands the button to occupy the full available width of its container.
   *
   * @default false
   */
  block?: boolean;

  /**
   * Callback executed when the button is clicked.
   *
   * Supports both synchronous and asynchronous handlers.
   */
  onClick?: (e?: MouseEvent) => void | Promise<void>;
}
