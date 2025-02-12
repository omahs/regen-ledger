Feature: MsgUpdateBasketFee

  Scenario: a valid message
    Given the message
    """
    {
      "authority": "regen1elq7ys34gpkj3jyvqee0h6yk4h9wsfxmgqelsw",
      "fee": {
        "denom": "uregen",
        "amount": "1000"
      }
    }
    """
    When the message is validated
    Then expect no error

  Scenario: a valid message removing basket fee
    Given the message
    """
    {
      "authority": "regen1elq7ys34gpkj3jyvqee0h6yk4h9wsfxmgqelsw"
    }
    """
    When the message is validated
    Then expect no error

  Scenario: an error is returned if authority is empty
    Given the message
    """
    {}
    """
    When the message is validated
    Then expect the error "invalid authority address: empty address string is not allowed"

  Scenario: an error is returned if authority is not a valid bech32 address
    Given the message
    """
    {
      "authority":"foo"
    }
    """
    When the message is validated
    Then expect the error "invalid authority address: decoding bech32 failed: invalid bech32 string length 3"

  Scenario: an error is returned if basket fee denom is not formatted
    Given the message
    """
    {
      "authority": "regen1elq7ys34gpkj3jyvqee0h6yk4h9wsfxmgqelsw",
      "fee": {
        "denom": "1",
        "amount": "1000"
      }
    }
    """
    When the message is validated
    Then expect the error "invalid denom: 1: invalid request"
