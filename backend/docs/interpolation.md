# Data Interpolation Process

This document describes the process of interpolating missing data points within datasets that represent measurements over time at fixed geographic locations. The goal is to fill in gaps (`NaN`) based on surrounding valid data points according to specific rules.

## Description

The core of the interpolation logic is the `interpolateLinearyDataRow` function. This function is designed to process a sequence of data points for a _single_ geographic location, ordered by their measurement time. It identifies missing values and attempts to fill them based on the presence and values of neighboring data points within that sequence.

`interpolateLinearyDataRow()` applies the following rules:

1.  **Direct Neighbors Interpolation:**

    - If a data point at a specific time `T` is missing, but the points at the immediately preceding time `T-1` and immediately succeeding time `T+1` for the _same location_ are valid, the missing value at `T` is filled with the simple average of the values at `T-1` and `T+1`.

2.  **Linear Interpolation for Longer Gaps:**

    - If a data point at time `T` is missing and the value at `T-1` is valid, the function searches forward in time for the _next_ valid data point at time `T+n` for the same location.
    - If such a point `T+n` is found, all missing values between `T` and `T+n-1` (inclusive) are filled using linear interpolation between the valid value at `T-1` and the valid value at `T+n`. The interpolated values are evenly distributed across the gap.

3.  **Unbounded Gaps Remain Unfilled:**
    - If a series of missing values occurs at the beginning or end of a location's data record, or if a gap is not bounded by valid data points on _both_ sides (either immediately or further down the sequence), those missing values will _not_ be interpolated and will remain as missing (`NaN`).

The `interpolateLinearyDataRow` function works with any data structure that implements the `InterpolatableData` interface, allowing it to be applied to various datasets within the project.

## `InterpolatableData` Interface

The `InterpolatableData` interface defines the contract for any data point that can be processed by the interpolation logic. It requires two methods:

```go
type InterpolatableData interface {
	Value() float32     // Returns the data value (e.g., chlorophyll concentration)
	SetValue(float32)   // Sets the data value
}
```

This interface allows `interpolateLinearyDataRow` to operate on the values without knowing the specific details of the underlying struct (like `models.ChlorophyllData`).

## How to Add New Data to be Interpolated

This section outlines the steps to enable interpolation for a new data source, assuming the data source has already been integrated into the project following the instructions in `adding_new_data.md`.

1.  **Implement Database Queries:**

    - In a file (`database/queries-source_name.go`), write SQL queries and corresponding Go functions within the `database.Service` implementation to:
      - Find all unique geographic locations for the new data source that have missing values within a given time range.
      - Retrieve all data points for a specific location, ordered by measurement time.
      - Update the data value for a specific record based on its unique identifier (ID).

2.  **Implement `InterpolatableData` Interface:**

    - Ensure the Go struct that holds the data for your new source implements the `InterpolatableData` interface. This involves adding the `Value()` and `SetValue(float32)` methods to the struct.

3.  **Implement Interpolation Method in `Interpolator`:**

    - In `internal/interpolator/interpolator.go`, add a new method (e.g., `RunSourceNameInterpolation`) to the `Interpolator` struct. This method will orchestrate the interpolation process for your specific data source:
      - Call the database function to find locations with gaps.
      - For each location, retrieve the data points using the database function.
      - Convert the retrieved data points into a slice of `InterpolatableData`.
      - Call the `interpolateLinearyDataRow()` function or other implemented interpolation function with this slice.
      - Iterate through the modified slice and update the database for any values that were filled.

4.  **Integrate into the Updater:**
    - In `internal/scheduler/updater.go` find the `update` method that handles the daily data process for your new source.
    - After the data for this source has been successfully downloaded and saved to the database, call the `RunSourceNameInterpolation` method you created in the previous step to perform the gap filling.

## How to Create New Interpolation Functions

While `interpolateLinearyDataRow` provides the current interpolation logic, you can extend the system by creating alternative interpolation functions.

1.  Define a new function that takes a slice of `InterpolatableData` (or a similar interface if your new method requires different capabilities) and applies your desired interpolation algorithm.
2.  Integrate this new function into the `Interpolator` (or a new interpolation service) and expose it through appropriate methods.
3.  Update the relevant data processing workflows to use your new interpolation function.
4.  Preferably implement tests to guarantee correct functioning of new method.
