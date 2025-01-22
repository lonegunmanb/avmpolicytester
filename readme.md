# Policy Test Helper

This project includes a test helper for testing Rego policy files using mock JSON data. The test helper is designed to walk through a specified directory, find `.rego` files, and run tests on them using `conftest`. The results are validated against expected outcomes defined in corresponding mock JSON files.

## Structure of Rego Policy Files

The test helper assumes the following structure for Rego policy files and their corresponding mock JSON files:

Each `.rego` file should have a corresponding `.mock.json` file that contains the mock data for testing.

```text
└── test
├── policy1
│   ├── policy1.mock.json
│   ├── policy1.rego
│   ├── policy2.mock.json
│   └── policy2.rego
└── policy2
    ├── policy1.mock.json
    ├── policy1.rego
    ├── policy2.mock.json
    └── policy2.rego

```

## Mock JSON File structure

The test helper function assumes that the mock JSON file has a specific structure. The mock JSON file should contain a top-level key named "mock", which maps to a dictionary. This dictionary can have keys `"valid"` and `"invalid"`, each mapping to another dictionary of test cases.

```json
{
  "mock": {
    "valid": {
      "case1": {...},
      "case2": {...}
    },
    "invalid": {
      "case1": {...},
      "case2": {...}
    }
  }
}
```

Or you can put all cases under `mock` key directly.

```json
{
  "mock": {
    "case1": {...},
    "invalid_case2": {...}
  }
}
```

Any keys other than `valid` and `invalid` would be treated as a single case, any single cases without `invalid` prefix would be considered as a valid case.

## Assigning Utility Rego Files

If you have utility Rego files that should be included in the tests, you can specify them using the UTILS_REGO environment variable. The UTILS_REGO variable should contain a comma-separated list of paths to the utility Rego files. For example:

```bash
export UTILS_REGO=/path/to/util1.rego,/path/to/util2.rego
```
