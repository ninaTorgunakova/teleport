import { Flex, Input } from 'design';

import React from 'react';

import { InputLabel } from 'design/Input/InputLabel';

import { RoleProps } from 'teleport/Welcome/Questionnaire/types';

export const Role = ({ team, role, updateFields }: RoleProps) => {
  return (
    <Flex gap="2%" mb={2}>
      <Flex flexDirection="column" width="49%">
        <InputLabel label="Which Team are you on?" aria="team" required />
        <Input
          id="team"
          type="text"
          value={team}
          placeholder="software eng"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateFields({ team: e.target.value });
          }}
        />
      </Flex>
      <Flex flexDirection="column" width="49%">
        <InputLabel
          label="What best describes your role?"
          aria="role"
          required
        />
        <Input
          id="role"
          type="text"
          value={role}
          placeholder="individual contributer"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateFields({ role: e.target.value });
          }}
        />
      </Flex>
    </Flex>
  );
};
