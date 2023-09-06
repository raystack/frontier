export type SecurityCheckboxTypes = {
  label: string;
  name: string;
  text: string;
  value: boolean;
  onValueChange: (key: string, checked: boolean) => void;
};
