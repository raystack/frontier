import { Button, Checkbox, Command, Flex, Popover } from "@raystack/apsara-v1";
import React from "react";

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
      <Popover.Trigger
        render={
          <Button data-test-id="admin-selected">
            {selectedValues.size} selected
          </Button>
        }
      />
      <Popover.Content align="start" style={{ padding: 0 }}>
        <Command>
          <Command.Input />
          <Command.Content>
            <Command.Empty>No results found.</Command.Empty>
            <Command.Group>
              {options.map((option: Option<T>) => {
                const isSelected = selectedValues.has(option.value);
                return (
                  <Command.Item
                    key={option.value as React.Key}
                    onClick={() => handleSelect(option.value)}
                    data-test-id={`frontier-admin-multiselect-item-${option.value}`}
                  >
                    <Flex
                      justify="start"
                      gap={3}
                      style={{ padding: "4px 0px", cursor: "pointer" }}
                    >
                      <Checkbox checked={isSelected} style={{ pointerEvents: "none" }} />
                      <Flex align="center" gap={3}>
                        <span>{option.label}</span>
                      </Flex>
                    </Flex>
                  </Command.Item>
                );
              })}
            </Command.Group>
          </Command.Content>
          <>
            <Command.Separator />
            <Command.Group>
              <Command.Item onClick={onClear} data-test-id="frontier-admin-multiselect-clear">
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
