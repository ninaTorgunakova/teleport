import React from 'react';

import { useTheme } from 'styled-components';

import {Flex, Text} from "design";

type InputLabelProps = {
  label: string;
  aria: string;
  required: boolean;
};

export const InputLabel = ({ label, aria, required }: InputLabelProps) => {
    const theme = useTheme();
    return (
        <Flex gap={1} mb={1}>
            <label
                htmlFor={aria}
                aria-label={aria}
                aria-required={required}
                style={{
                    fontWeight: 300,
                    fontSize: '14px',
                    lineHeight: '20px',
                }}
                data-testid="aria"
            >
                {label}
            </label>
            {required && <Text color={theme.colors.error.main}>*</Text>}
        </Flex>
    );
}
