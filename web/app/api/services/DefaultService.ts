/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AddContactRequest } from '../models/AddContactRequest';
import type { AddContactResponse } from '../models/AddContactResponse';
import type { AuthRequest } from '../models/AuthRequest';
import type { AuthResponse } from '../models/AuthResponse';
import type { DeleteContactResponse } from '../models/DeleteContactResponse';
import type { GetContactRequest } from '../models/GetContactRequest';
import type { GetContactResponse } from '../models/GetContactResponse';
import type { GetContactsRequest } from '../models/GetContactsRequest';
import type { GetContactsResponse } from '../models/GetContactsResponse';
import type { RegisterRequest } from '../models/RegisterRequest';
import type { RegisterResponse } from '../models/RegisterResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DefaultService {
    /**
     * User authentication
     * @param requestBody
     * @returns AuthResponse Authentication successful
     * @throws ApiError
     */
    public static postApiV1Auth(
        requestBody: AuthRequest,
    ): CancelablePromise<AuthResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/v1/auth',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request`,
                401: `Invalid login or password`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Register new user
     * @param requestBody
     * @returns RegisterResponse Registration successful
     * @throws ApiError
     */
    public static postApiV1AuthRegister(
        requestBody: RegisterRequest,
    ): CancelablePromise<RegisterResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/v1/auth/register',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request`,
                409: `User already exists`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Retrieve a list of contacts
     * @param requestBody
     * @returns GetContactsResponse Contacts retrieved successfully
     * @throws ApiError
     */
    public static postApiV1Contacts(
        requestBody: GetContactsRequest,
    ): CancelablePromise<GetContactsResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/v1/contacts',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request`,
                401: `Unauthorized`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Add a new contact
     * @param requestBody
     * @returns AddContactResponse Contact added successfully
     * @throws ApiError
     */
    public static postApiV1Contact(
        requestBody: AddContactRequest,
    ): CancelablePromise<AddContactResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/v1/contact',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request`,
                401: `Unauthorized`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Retrieve a contact
     * @param contactId
     * @param requestBody
     * @returns GetContactResponse Contact retrieved successfully
     * @throws ApiError
     */
    public static postApiV1Contact1(
        contactId: number,
        requestBody: GetContactRequest,
    ): CancelablePromise<GetContactResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/v1/contact/{contact_id}',
            path: {
                'contact_id': contactId,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request`,
                401: `Unauthorized`,
                404: `Contact not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Delete a contact
     * @param contactId
     * @returns DeleteContactResponse Contact deleted successfully
     * @throws ApiError
     */
    public static deleteApiV1Contact(
        contactId: number,
    ): CancelablePromise<DeleteContactResponse> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/api/v1/contact/{contact_id}',
            path: {
                'contact_id': contactId,
            },
            errors: {
                400: `Invalid request`,
                401: `Unauthorized`,
                404: `Contact not found`,
                500: `Internal server error`,
            },
        });
    }
}
