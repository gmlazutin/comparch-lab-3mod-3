/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AddPhoneObject } from './AddPhoneObject';
import type { NonEmptyString } from './NonEmptyString';
export type AddContactRequest = {
    name: NonEmptyString;
    birthday: string;
    note?: string;
    initialPhones: Array<AddPhoneObject>;
};

