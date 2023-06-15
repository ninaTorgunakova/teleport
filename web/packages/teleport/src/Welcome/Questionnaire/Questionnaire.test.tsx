import { screen } from '@testing-library/react';

import React from 'react';

import { render } from 'design/utils/testing';

import { Questionnaire } from 'teleport/Welcome/Questionnaire/Questionnaire';

describe('questionnaire', () => {
  test('loads each question', () => {
    render(<Questionnaire />);

    expect(screen.getByText('Tell us about yourself')).toBeInTheDocument();
    expect(
      screen.getByText(
        'We will customize your Teleport experience based on your choices.'
      )
    ).toBeInTheDocument();

    expect(screen.getByLabelText('Company Name')).toBeInTheDocument();
    expect(screen.getByLabelText('Number of Employees')).toBeInTheDocument();
    expect(screen.getByLabelText('Which Team are you on?')).toBeInTheDocument();
    expect(
      screen.getByLabelText('What best describes your role?')
    ).toBeInTheDocument();
    expect(
      screen.getByLabelText(
        'Which infrastructure resources do you need to access frequently? Select all that apply.'
      )
    ).toBeInTheDocument();
  });

  test('disables submit until required fields are inputed', () => {
    render(<Questionnaire />);

    expect(screen.getByRole('button', { name: 'Submit' })).toBeDisabled();

    //fireEvent.change(input, { target: { value: 'new-email@hey.com' } });
  });
});
