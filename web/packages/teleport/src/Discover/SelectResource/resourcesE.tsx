import { DiscoverEventResource } from 'teleport/services/userEvent';

import { ResourceKind } from '../Shared';

import { ResourceSpec } from './types';

export const SAML_APPLICATIONS: ResourceSpec[] = [
  {
    name: 'SAML Application',
    kind: ResourceKind.SamlApplication,
    keywords: 'saml sso application idp',
    icon: 'Application',
    event: DiscoverEventResource.SamlApplication,
  },
  {
    name: 'SAML Application (Grafana)',
    kind: ResourceKind.SamlApplication,
    keywords: 'saml sso application idp grafana',
    icon: 'Grafana',
    event: DiscoverEventResource.SamlApplication,
  },
];
