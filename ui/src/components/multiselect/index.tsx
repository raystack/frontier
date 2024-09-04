import { Button, Checkbox, Command, Flex, Popover } from "@raystack/apsara";

interface Option<T> {
  value: T;
  label: string;
}

interface MultiSelectProps<T> {
  selected: T[];
  options: Option<T>[];
  onSelect: (values: T[]) => void;
}

export function MultiSelect<T>({
  selected,
  options = [],
  onSelect = () => {},
}: MultiSelectProps<T>) {
  const selectedValues = new Set(selected);

  function handleSelect(value: T) {
    if (selectedValues.has(value)) {
      selectedValues.delete(value);
    } else {
      selectedValues.add(value);
    }

    const selectedValuesArr = Array.from(selectedValues);
    onSelect(selectedValuesArr);
  }

  function onClear() {
    onSelect([]);
  }

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button data-test-id="admin-ui-selected">{selectedValues.size} selected</Button>
      </Popover.Trigger>
      <Popover.Content align="start" style={{ padding: 0 }}>
        <Command>
          <Command.Input />
          <Command.List>
            <Command.Empty>No results found.</Command.Empty>
            <Command.Group>
              {options.map((option: Option<T>) => {
                const isSelected = selectedValues.has(option.value);
                return (
                  <Command.Item
                    key={option.value}
                    onSelect={() => handleSelect(option.value)}
                  >
                    <Flex
                      justify="start"
                      gap="small"
                      style={{ padding: "4px 0px", cursor: "pointer" }}
                    >
                      <Checkbox checked={isSelected} />
                      <Flex align="center" gap="small">
                        <span>{option.label}</span>
                      </Flex>
                    </Flex>
                  </Command.Item>
                );
              })}
            </Command.Group>
          </Command.List>
          <>
            <Command.Separator />
            <Command.Group>
              <Command.Item onSelect={onClear}>
                <Flex
                  justify="center"
                  style={{ padding: "4px 0px", cursor: "pointer" }}
                >
                  Clear
                </Flex>
              </Command.Item>
            </Command.Group>
          </>
        </Command>
      </Popover.Content>
    </Popover>
  );
}
