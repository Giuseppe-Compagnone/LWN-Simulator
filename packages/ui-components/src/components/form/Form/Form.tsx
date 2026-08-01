import { Button } from "@/components/common";
import { FormProps } from "./Form.types";
import { useForm } from "./useForm";
import cn from "classnames";

const Form = (props: FormProps) => {
  const formLogic = useForm({ ...props });

  return (
    <form className="form" noValidate>
      {Object.values(formLogic.fieldsState).map((field) => (
        <div className="form-field-wrapper" key={field.name}>
          <div className="form-field-header">
            <label className="field-label" htmlFor={field.name}>
              {field.label}
              {field.required && "*"}
            </label>
            <div className="error">{field.error}</div>
          </div>

          {field.render({
            ...field,
            value: formLogic.fieldsState[field.name].value,
            setValue: (v) => formLogic.setValue(field.name, v),
            disabled: formLogic.isFieldDisabled(field, formLogic.fieldsState),
          })}
          {field.info && (
            <div className="field-info">
              <span className="material-symbols-outlined icon">info</span>
              <span>{field.info}</span>
            </div>
          )}
        </div>
      ))}
      <Button
        {...(props.submitButton || { value: "Submit" })}
        className={cn("form-button", props.submitButton?.className)}
        onClick={(e) => {
          e?.preventDefault();
          const isValid = formLogic.validate();
          if (isValid)
            props.onSubmit(
              Object.fromEntries(
                Object.values(formLogic.fieldsState)
                  .filter(
                    (field) =>
                      !formLogic.isFieldDisabled(field, formLogic.fieldsState),
                  )
                  .map((field) => [field.name, field.value]),
              ),
            );
        }}
        disabled={
          !!Object.values(formLogic.fieldsState).find(
            (field) =>
              !formLogic.isFieldDisabled(field, formLogic.fieldsState) &&
              field.error,
          )
        }
      />
    </form>
  );
};

export default Form;
