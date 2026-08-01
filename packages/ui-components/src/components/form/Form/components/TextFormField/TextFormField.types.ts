import { FormField, FormFieldProps } from "../..";

/**
 * Properties for configuring a text-based form field.
 *
 * Extends the base form field properties with text-specific behavior.
 */
export interface TextFormFieldProps extends FormFieldProps {
  /**
   * Current text value of the field.
   */
  value: string;

  /**
   * Updates the current text value.
   *
   * @param value New text value assigned to the field.
   */
  setValue(value: string): void;

  /**
   * Indicates whether the field value should be displayed using a mask.
   *
   * Useful for hiding sensitive text values such as passwords.
   *
   * @default false
   */
  masked?: boolean;
}

/**
 * Configuration options for creating a text form field.
 *
 * Excludes rendering logic and value update handling, which are managed
 * internally by the form system.
 */
export interface TextFieldOptions extends Omit<
  TextFormFieldProps,
  "render" | "setValue" | "disabled"
> {
  disabled?: boolean | ((fieldsState: Record<string, FormField>) => boolean);
}
