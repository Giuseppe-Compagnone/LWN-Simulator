import { testCompProps } from "./testComp.types";

const testComp = (props: testCompProps) => {
  return <h1>Test Comp {props.index}</h1>;
};

export default testComp;
