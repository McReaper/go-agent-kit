Feature: Notes
  Someone keeps short notes, from the command line or through an agent over MCP.

  Scenario: A note added from the command line is in the list
    When I add the note "Groceries" with the text "eggs, milk"
    Then the list shows the note "Groceries"

  Scenario: A title over 80 characters is refused
    When I add a note whose title is 81 characters long
    Then it is refused with "the limit is 80"

  Scenario: An unknown option is a usage error, not a refusal
    When I run "add-note --no-such-flag"
    Then it exits with code 2

  Scenario: An agent adds a note over MCP, then finds it in the list
    When an agent adds the note "Groceries"
    Then the agent's list shows the note "Groceries"

  @wip
  Scenario: A deleted note leaves the list
    Given the note "Groceries"
    When I delete the note "Groceries"
    Then the list shows 0 notes
