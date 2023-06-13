import { Flex, Text } from 'design';

import React from 'react';

import { InputLabel } from 'design/Input/InputLabel';

import Image from 'design/Image';

import { useTheme } from 'styled-components';

import { CheckboxInput } from 'design/Checkbox';

import {
  ResourcesProps,
  ResourceType,
} from 'teleport/Welcome/Questionnaire/types';

export const Resources = ({
  resources,
  checked,
  updateFields,
}: ResourcesProps) => {
  const theme = useTheme();

  const updateResources = (label: string) => {
    let updated = checked;
    if (updated.includes(label)) {
      updated = updated.filter(r => r !== label);
    } else {
      updated.push(label);
    }

    updateFields({ resources: updated });
  };

  const renderCheck = (resource: ResourceType, index: number) => {
    return (
      <Flex
        key={`${index}-${resource.label}`}
        flexDirection="column"
        width="20%"
        height="100%"
        bg={theme.colors.spotBackground[0]}
        p="16px 0"
        justifyContent="space-between"
      >
        <Flex
          flexDirection="column"
          alignItems="center"
          justifyContent="center"
          height="100%"
        >
          <Image src={resource.image} height="64px" width="64px" />
          <Text textAlign="center" typography="body3">
            {resource.label}
          </Text>
        </Flex>
        <CheckboxInput
          type="checkbox"
          name={resource.image}
          id={resource.image}
          onChange={() => {
            updateResources(resource.label);
          }}
          checked={checked.includes(resource.label)}
        />
      </Flex>
    );
  };

  return (
    <>
      <InputLabel
        label="Which infrastructure resources do you need to access frequently?"
        aria="resources"
        required
      />
      <Flex gap={2} alignItems="flex-start" height="170px">
        {resources.map((r: ResourceType, i: number) => renderCheck(r, i))}
      </Flex>
    </>
  );
};
