export type QuestionProps = {
  updateFields: (fields: Partial<QuestionnaireFormFields>) => void;
};

export type CompanyProps = QuestionProps & {
  companyName: string;
  numberOfEmployees: EmployeeRangeStrings;
};

export type RoleProps = QuestionProps & {
  role: string;
  team: string;
};

export type ResourceType = {
  label: string;
  image: string;
};

export type ResourcesProps = QuestionProps & {
  resources: ResourceType[];
  checked: string[];
};

export type QuestionnaireFormFields = {
  companyName: string;
  employeeCount: EmployeeRangeStrings;
  role: string;
  team: string;
  resources: string[];
};

export enum EmployeeRange {
  '0-100',
  '100+',
}

export type EmployeeRangeStrings = keyof typeof EmployeeRange;
