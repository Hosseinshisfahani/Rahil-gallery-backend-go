Feature: Customer authentication
  As a jewelry store customer
  I want to register, sign in, and manage my session
  So that I can access my account securely

  Background:
    Given the auth API is running

  Scenario: Register with valid credentials
    Given no user exists with email "bdd-user@example.com"
    When I register with:
      | email      | bdd-user@example.com |
      | password   | Secret12             |
      | first_name | Sara                 |
      | last_name  | Rahil                |
    Then the response status should be 201
    And the response should contain access and refresh tokens

  Scenario: Cannot register with weak password
    When I register with:
      | email      | weak@example.com |
      | password   | short            |
      | first_name | A                |
      | last_name  | B                |
    Then the response status should be 400
    And the error code should be "validation_error"

  Scenario: Login and view profile
    Given a registered user:
      | email    | login-bdd@example.com |
      | password | Secret12              |
    When I login with email "login-bdd@example.com" and password "Secret12"
    Then the response status should be 200
    And the response should contain access and refresh tokens
    When I request my profile with the saved access token
    Then the response status should be 200
    And the profile email should be "login-bdd@example.com"

  Scenario: Refresh token rotation
    Given a registered user:
      | email    | refresh-bdd@example.com |
      | password | Secret12                |
    When I login with email "refresh-bdd@example.com" and password "Secret12"
    And I refresh the session with the saved refresh token
    Then the response status should be 200
    And the refresh token should be rotated

  Scenario: Logout invalidates refresh token
    Given a registered user:
      | email    | logout-bdd@example.com |
      | password | Secret12               |
    When I login with email "logout-bdd@example.com" and password "Secret12"
    And I logout with the saved refresh token
    Then the response status should be 200
    When I refresh the session with the saved refresh token
    Then the response status should be 401
    And the error code should be "invalid_refresh_token"

  Scenario: Profile requires authentication
    When I request my profile without a token
    Then the response status should be 401
