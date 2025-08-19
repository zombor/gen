# Project: Gen

## Testing Style

- Always use `ginkgo` and `gomega` for testing
- Use one assertion per `It` block
- If a function returns a value with an error, check the error with `Expect(foo, err).To(Equal(1))` style instead of `Expect(err).ToNot(HaveOccurred())` style
- Use `BeforeEach` to setup test state
- Use `JustBeforeEach` to invoke the subject under test
- Tests should cover 100% of the code under test

## Coding Style

- To make testing dependencies easier, use a functional injection style. Inject functions of dependencies into business logic so that we can mock them in tests.

## Regarding Dependencies:

- Avoid introducing new external dependencies unless absolutely necessary.
- If a new dependency is required, please state the reason.