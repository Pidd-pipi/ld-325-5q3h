package errors

type PriceNotFoundError struct{ ProductID uint }

func (e *PriceNotFoundError) Error() string { return "price quote not found" }
