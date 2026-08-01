import { LogoProps } from "./Logo.types";

const Logo = (props: LogoProps) => {
  return (
    <div className="logo">
      <img src={"./logo.svg"} />
    </div>
  );
};

export default Logo;
