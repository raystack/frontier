export type SecurityCheckboxTypes = {
  label: string;
  name: string;
  text: string;
  value: boolean;
  canUpdatePreference?: boolean;
  onValueChange: (key: string, checked: boolean) => void;
};
