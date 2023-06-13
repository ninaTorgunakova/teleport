import React, {useState} from 'react';
import {ButtonPrimary, Card, Text} from 'design';
import {useTheme} from 'styled-components';

import {QuestionnaireFormFields} from 'teleport/Welcome/Questionnaire/types';
import {Company} from 'teleport/Welcome/Questionnaire/Company';
import {Role} from 'teleport/Welcome/Questionnaire/Role';
import {Resources} from 'teleport/Welcome/Questionnaire/Resources';

// todo mberg move these assets
import application from '../../components/Empty/assets/appplication.png';
import desktop from '../../components/Empty/assets/desktop.png';
import database from '../../components/Empty/assets/database.png';
import stack from '../../components/Empty/assets/stack.png';

export const Questionnaire = () => {
    const theme = useTheme();
    const supportedResources = [
        {label: 'Web Applications', image: application},
        {label: 'Windows Desktops', image: desktop},
        {label: 'Server/SSH', image: stack},
        {label: 'Databases', image: database},
        {label: 'Kubernetes', image: stack},
    ];
    const [formFields, setFormFields] = useState<QuestionnaireFormFields>({
        companyName: '',
        employeeCount: '0-100',
        team: '',
        role: '',
        resources: []
    });

    const updateForm = (fields: Partial<QuestionnaireFormFields>) => {
        setFormFields({
            role: fields.role ?? formFields.role,
            team: fields.team ?? formFields.team,
            resources: fields.resources ?? formFields.resources,
            companyName: fields.companyName ?? formFields.companyName,
            employeeCount: fields.employeeCount ?? formFields.employeeCount,
        });
    };

    const submitForm = () => {
        console.log(formFields);
        // submit to storage
        // submit ph event
        // redirect to web/discover
    };

    return (
        <Card mx="auto" maxWidth="600px" p="4">
            <Text typography="h2">Tell us about yourself</Text>
            <Text typography="subtitle1">
                We will customize your Teleport experience based on your choices.
            </Text>
            <hr
                style={{
                    margin: '24px 0',
                    border: `1px solid ${theme.colors.spotBackground[0]}`,
                }}
            />
            <Company
                companyName={formFields.companyName}
                numberOfEmployees={formFields.employeeCount}
                updateFields={updateForm}
            />
            <Role
                role={formFields.role}
                team={formFields.team}
                updateFields={updateForm}
            />
            <Resources resources={supportedResources} checked={formFields.resources} updateFields={updateForm}/>
            <ButtonPrimary
                mt={3}
                width="100%"
                size="large"
                disabled={false}
                onClick={submitForm}
            >
                Submit
            </ButtonPrimary>
        </Card>
    );
};
