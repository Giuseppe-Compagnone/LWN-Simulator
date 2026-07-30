export enum FormFieldType {
  Text = 1,
}

export interface FormField {
  type: FormFieldType;
  name: string;
  label: string;
  value: string;
}

export interface FormProps {
  fields: Array<FormField>;
}

export interface UseFormProps extends FormProps {}

export interface FormLogic {
  fields: Array<FormField>;
  setFields: (fields: Array<FormField>) => void;
}

export interface FormFieldComponent {
  field: FormField;
}
