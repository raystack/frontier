export type SecurityCheckboxTypes = {
  label: string;
  name: string;
  text: string;
  value: boolean;
  canUpdatePrefrence?: boolean;
  onValueChange: (key: string, checked: boolean) => void;
};
