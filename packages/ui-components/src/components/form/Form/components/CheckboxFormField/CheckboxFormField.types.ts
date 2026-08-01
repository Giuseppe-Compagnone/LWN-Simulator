import { JSX } from "react/jsx-runtime";
import { FormField, FormFieldProps } from "../..";

/**
 * Represents an option available in a checkbox form field.
 */
export interface CheckboxOption {
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
 * Properties for configuring a checkbox-based form field.
 *
 * Extends the base form field properties with multi-selection behavior.
 */
export interface CheckboxFormFieldProps extends FormFieldProps {
  /**
   * Currently selected option values.
   *
   * A `null` value indicates that no option is currently selected.
   */
  value: Array<string> | null;

  /**
   * Updates the selected option values.
   *
   * @param value New selected values, or `null` to clear the selection.
   */
  setValue(value: Array<string> | null): void;

  /**
   * Available options displayed by the checkbox group.
   */
  options: Array<CheckboxOption>;

  /**
   * Checkbox fields do not support placeholder text.
   */
  placeholder?: undefined;
  format?: undefined;
  validations?: undefined;
}

/**
 * Configuration options for creating a checkbox form field.
 *
 * Excludes rendering logic and value update handling, which are managed
 * internally by the form system.
 */
export interface CheckboxFieldOptions extends Omit<
  CheckboxFormFieldProps,
  "render" | "setValue" | "disabled"
> {
  disabled?: boolean | ((fieldsState: Record<string, FormField>) => boolean);
}
