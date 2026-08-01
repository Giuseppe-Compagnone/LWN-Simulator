import { ButtonProps } from "@/components/common";
import { ReactNode } from "react";

/**
 * Represents the possible values supported by form fields.
 */
export type FormValue = string | number | boolean | Date | null | Array<string>;

/**
 * Properties passed to a custom form field renderer.
 *
 * Extends the base field configuration while providing access to the
 * value update callback.
 */
export interface FormFieldProps extends Omit<FormField, "render"> {
  /**
   * Updates the value of the current field.
   *
   * @param value New value assigned to the field.
   */
  setValue(value: FormValue): void;
}

/**
 * Configuration of a single form field.
 */
export interface FormField {
  /**
   * Unique identifier of the field.
   */
  name: string;

  /**
   * Label displayed for the field.
   */
  label: string;

  /**
   * Current value assigned to the field.
   */
  value: FormValue;

  /**
   * Placeholder text displayed when the field has no value.
   */
  placeholder?: string;

  /**
   * Additional information or helper text displayed alongside the field.
   */
  info?: string;

  format?: (raw: string) => string;

  error: string | null;

  required?: boolean;

  disabled?: boolean;

  active?: boolean;

  /**
   * Custom renderer used to define the field UI.
   *
   * Receives the field properties and the callback required to update
   * its value.
   */
  render(props: FormFieldProps): ReactNode;
}

/**
 * Properties for configuring a form component.
 */
export interface FormProps {
  /**
   * Collection of fields rendered by the form.
   */
  fields: FormField[];

  /**
   * Callback executed when the form is submitted.
   *
   * Receives an object containing the current values indexed by field name.
   *
   * Supports both synchronous and asynchronous handlers.
   */
  onSubmit(values: Record<string, FormValue>): void | Promise<void>;

  /**
   * Configuration of the submit button displayed by the form.
   */
  submitButton?: ButtonProps;
}

/**
 * Properties accepted by the form logic hook.
 *
 * Extends the form configuration properties.
 */
export interface UseFormProps extends FormProps {}

/**
 * State and actions exposed by the form logic hook.
 */
export interface FormLogic {
  fieldsState: Record<string, FormField>;
  /**
   * Updates the value of a specific field.
   *
   * @param name Name of the field to update.
   * @param value New value assigned to the field.
   */
  setValue(name: string, value: FormValue): void;
  validate: () => boolean;
}
