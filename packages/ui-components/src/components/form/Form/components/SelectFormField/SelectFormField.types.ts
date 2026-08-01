import { JSX } from "react/jsx-runtime";
import { FormField, FormFieldProps } from "../..";

/**
 * Represents an option available in a select form field.
 */
export interface SelectOption {
  /**
   * Value assigned to the option.
   */
  value: string;

  /**
   * Custom content displayed instead of the default option value.
   */
  displayed?: JSX.Element;
}

/**
 * Properties for configuring a select-based form field.
 *
 * Extends the base form field properties with selection-specific behavior.
 */
export interface SelectFormFieldProps extends FormFieldProps {
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
   * Available options displayed by the select field.
   */
  options: Array<SelectOption>;
  format?: undefined;
  validations?: undefined;
}

/**
 * Configuration options for creating a select form field.
 *
 * Excludes rendering logic and value update handling, which are managed
 * internally by the form system.
 */
export interface SelectFieldOptions extends Omit<
  SelectFormFieldProps,
  "render" | "setValue" | "disabled"
> {
  disabled?: boolean | ((fieldsState: Record<string, FormField>) => boolean);
}
