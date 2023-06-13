import { Flex, Input } from 'design';

import React, { useMemo } from 'react';

import Select, { Option } from 'shared/components/Select';

import { InputLabel } from 'design/Input/InputLabel';

import {
  CompanyProps,
  EmployeeRange,
  EmployeeRangeStrings,
} from 'teleport/Welcome/Questionnaire/types';

export const Company = ({
  updateFields,
  companyName,
  numberOfEmployees,
}: CompanyProps) => {
  const options: Option<EmployeeRangeStrings, EmployeeRangeStrings>[] =
    useMemo(() => {
      return Object.values(EmployeeRange)
        .filter(v => !isNaN(Number(v)))
        .map(key => ({
          value: EmployeeRange[key],
          label: EmployeeRange[key],
        }));
    }, []);

  return (
    <Flex gap="2%" mb={2}>
      <Flex flexDirection="column" width="49%">
        <InputLabel label="Company Name" aria="company-name" required />
        <Input
          id="company-name"
          type="text"
          value={companyName}
          placeholder="ex. github"
          onChange={e => {
            updateFields({
              companyName: e.target.value,
            });
          }}
        />
      </Flex>
      <Flex flexDirection="column" width="49%">
        <InputLabel label="Number of Employees" aria="employees" required />
        <Select
          inputId="employees"
          label="employees"
          hasError={false}
          elevated={false}
          onChange={(e: Option<EmployeeRangeStrings, string>) =>
            updateFields({
              employeeCount: e.value,
            })
          }
          options={options}
          value={{
            value: numberOfEmployees,
            label: numberOfEmployees,
          }}
        />
      </Flex>
    </Flex>
  );
};
