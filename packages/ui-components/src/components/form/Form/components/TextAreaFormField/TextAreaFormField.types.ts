import { FormField, FormFieldProps } from "../..";

/**
 * Properties for configuring a textarea-based form field.
 *
 * Extends the base form field properties with multiline text-specific behavior.
 */
export interface TextAreaFormFieldProps extends FormFieldProps {
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
   * Enables resizing of the textarea element.
   *
   * When enabled, the user can manually adjust the textarea vertically.
   *
   * @default false
   */
  resize?: boolean;

  /**
   * Maximum number of characters allowed in the field.
   *
   * When provided, input exceeding this limit should be prevented
   */
  charsMax?: number;
}

/**
 * Configuration options for creating a textarea form field.
 *
 * Excludes rendering logic and value update handling, which are managed
 * internally by the form system.
 */
export interface TextAreaFieldOptions extends Omit<
  TextAreaFormFieldProps,
  "render" | "setValue" | "disabled"
> {
  disabled?: boolean | ((fieldsState: Record<string, FormField>) => boolean);
}
