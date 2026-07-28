export enum SpinnerSize {
  Sm = "0.4px",
  Md = "0.8px",
  Lg = "1.2px",
}

export enum SpinnerType {
  Primary = "--primary-color",
  Secondary = "--secondary-color",
}

export interface SpinnerProps {
  size?: SpinnerSize;
  type?: SpinnerType;
}
