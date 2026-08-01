import { JSX } from "react/jsx-runtime";
import { FormFieldProps } from "../..";

/**
 * Represents an option available in a radio form field.
 */
export interface RadioOption {
  /**
   * Value assigned to the option.
   */
  value: string;

  /**
   * Custom content displayed instead of the default option label.
   */
  displayed?: JSX.Element;
}

/**
 * Properties for configuring a radio-based form field.
 *
 * Extends the base form field properties with radio selection-specific behavior.
 */
export interface RadioFormFieldProps extends FormFieldProps {
  /**
   * Currently selected option value.
   *
   * A `null` value indicates that no option is currently selected.
   */
  value: string | null;

  /**
   * Updates the selected option value.
   *
   * @param value New selected value, or `null` to clear the selection.
   */
  setValue(value: string | null): void;

  /**
   * Available options displayed by the radio group.
   */
  options: Array<RadioOption>;

  /**
   * Radio fields do not support placeholder text.
   */
  placeholder?: undefined;
  format?: undefined;
}

/**
 * Configuration options for creating a radio form field.
 *
 * Excludes rendering logic and value update handling, which are managed
 * internally by the form system.
 */
export interface RadioFieldOptions extends Omit<
  RadioFormFieldProps,
  "render" | "setValue"
> {}
