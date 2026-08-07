"use client";

import { useState } from "react";
import { FormField, FormLogic, FormValue, UseFormProps } from "./Form.types";

export const useForm = (props: UseFormProps): FormLogic => {
  // States
  const [fieldsState, setFieldsState] = useState<Record<string, FormField>>(
    () => Object.fromEntries(props.fields.map((field) => [field.name, field])),
  );

  // Functions
  const setValue = (name: string, value: FormValue) => {
    setFieldsState((prev) => {
      const field = prev[name];
      field.value = value;
      field.error = null;

      if (field.validations && field.value) {
        for (let i = 0; i < field.validations.length; i++) {
          const validation = field.validations[i];

          if (!validation.realtime) continue;

          const isValid = validation.rule.test(field.value as string);
          if (!isValid) {
            field.error = validation.error;
            break;
          }
        }
      }

      return { ...prev };
    });
  };

  const isFieldDisabled = (
    field: FormField,
    fieldsState: Record<string, FormField>,
  ) => {
    return typeof field.disabled === "function"
      ? field.disabled(fieldsState)
      : !!field.disabled;
  };

  const isFieldDisplayed = (
    field: FormField,
    fieldsState: Record<string, FormField>,
  ) => {
    if (typeof field.display === "function") {
      return field.display(fieldsState);
    }

    return field.display ?? true;
  };

  const validate = () => {
    const tmp = { ...fieldsState };
    let isValid = true;

    Object.values(tmp).forEach((field) => {
      if (
        !isFieldDisabled(field, fieldsState) &&
        isFieldDisplayed(field, fieldsState)
      ) {
        if (field.required && !field.value) {
          isValid = false;
          field.error = "This field is required";
          console.log(field.name, field.error);
        }

        if (field.validations && field.value && isValid) {
          for (let i = 0; i < field.validations.length; i++) {
            const validation = field.validations[i];

            isValid = validation.rule.test((field.value as string) || "");
            if (!isValid) {
              field.error = validation.error;
              break;
            }
          }
        }
      }
    });

    if (!isValid) setFieldsState(tmp);

    return isValid;
  };

  return {
    fieldsState,
    setValue,
    isFieldDisabled,
    isFieldDisplayed,
    validate,
  };
};
