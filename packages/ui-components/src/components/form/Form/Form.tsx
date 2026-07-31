import { Button } from "@/components/common";
import { FormProps } from "./Form.types";
import { useForm } from "./useForm";
import cn from "classnames";

const Form = (props: FormProps) => {
  const formLogic = useForm({ ...props });

  return (
    <form className="form" noValidate>
      {props.fields.map((field) => (
        <div className="form-field-wrapper" key={field.name}>
          <label className="field-label" htmlFor={field.name}>
            {field.label}
          </label>

          {field.render({
            ...field,
            value: formLogic.values[field.name],
            setValue: (v) => formLogic.setValue(field.name, v),
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
        {...(props.button || { value: "Submit" })}
        className={cn("form-button", props.button?.className)}
        onClick={(e) => {
          e?.preventDefault();
          props.onSubmit(formLogic.values);
        }}
      />
    </form>
  );
};

export default Form;
