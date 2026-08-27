"use client";

import { useMemo, useRef, useState } from "react";
import { FormField, FormLogic, FormValue, UseFormProps } from "./Form.types";

export const useForm = (props: UseFormProps): FormLogic => {
  // States
  const [fieldsState, setFieldsState] = useState<Record<string, FormField>>(
    () => Object.fromEntries(props.fields.map((field) => [field.name, field])),
  );

  // Refs
  const fieldsStateRef = useRef<Record<string, FormField>>(
    Object.fromEntries(props.fields.map((field) => [field.name, field])),
  );

  // Functions
  const setValue = (name: string, value: FormValue) => {
    const prev = fieldsStateRef.current;
    const field = prev[name];

    const updatedField: FormField = {
      ...field,
      value,
      error: null,
    };

    if (updatedField.validations && updatedField.value) {
      for (const validation of updatedField.validations) {
        if (!validation.realtime) continue;

        const isValid = validation.rule.test(updatedField.value as string);

        if (!isValid) {
          updatedField.error = validation.error;
          break;
        }
      }
    }

    const nextState = {
      ...prev,
      [name]: updatedField,
    };

    fieldsStateRef.current = nextState;
    setFieldsState(nextState);

    if (updatedField.onChange) {
      const updatedLogic: FormLogic = {
        ...formLogic,
        fieldsState: nextState,
      };

      updatedField.onChange(updatedLogic);
    }
  };

  const setProp = (name: string, prop: string, value: unknown) => {
    const prev = fieldsStateRef.current;
    const field = prev[name];

    if (!field) return;

    const updatedField: FormField = {
      ...field,
      [prop]: value,
    };

    const nextState = {
      ...prev,
      [name]: updatedField,
    };

    fieldsStateRef.current = nextState;
    setFieldsState(nextState);
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

  const formLogic = useMemo<FormLogic>(() => {
    return {
      fieldsState,
      setValue,
      setProp,
      isFieldDisabled,
      isFieldDisplayed,
      validate,
    };
  }, [
    fieldsState,
    setValue,
    setProp,
    isFieldDisabled,
    isFieldDisplayed,
    validate,
  ]);

  return formLogic;
};
