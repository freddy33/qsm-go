# Simulation configs

`<simulation name>-config.json` is a minimal `inputs.json`. `qsm <name> --new` (or `./run run <name> --new`) reads `simulation-configs/<name>-config.json`, fills every omitted field, and writes the result to `data/<name>/inputs.json`. The written file keeps shortcuts. It does not replace a pyramid with the four events that shortcut expands to. The run file still records that expansion.

`QSM_CONFIG_ROOT` overrides this directory. `./run` sets it to `simulation-configs/` at the repository root. A relative value is resolved from the directory you started in.

If `data/<name>/` already exists, `--new` renames it to `data/<name>-<YYYYMMDDTHHMMSSZ>/` (UTC) before creating the new folder. A second archive in the same second is `data/<name>-<stamp>_01/`. A missing config is an error and the existing folder stays put.

## Defaults

Omitted keys use these values. They are the playgl world (`MakeWorld` and `CreatePyramidWithParams` with size 4).

| field | default |
| --- | --- |
| `max_time_steps` | `3` |
| `doc` | the playgl note, when the rest of the file is that default pyramid; otherwise empty |
| `space.active_path_node_threshold` | `3` |
| `space.max_trios_per_point` | `2` |
| `space.max_path_nodes_per_point` | `4` |
| `space.max_coord` | `18` |
| `events` | none |
| `shortcuts` | one default pyramid, when the key is absent |
| `pyramid.size` | `4` |
| `pyramid.growth_type` | `8` |
| `pyramid.growth_index` | `0` |
| `pyramid.offsets` | `[0, 2, 4, 6]` |
| `pyramid.creation_time` | `0` |
| event `x`, `y`, `z` | `0` |
| event `creation_time` | `0` |
| event `color` | `1` (red) |
| event `growth_type` | `8` |
| event `growth_index` | `0` |
| event `growth_offset` | `0` |

`"shortcuts": []` means no shortcuts. A missing `shortcuts` key means the default pyramid. Unknown keys are an error. `color` is `1` red, `2` green, `4` blue, `8` yellow. `max_coord` has to be a positive multiple of 3.

## Files

| file | simulation | what it grows |
| --- | --- | --- |
| `pyramid-config.json` | `pyramid` | the default pyramid used by the tests: size 4, through time 3 |
| `pyramid-size2-config.json` | `pyramid-size2` | the same pyramid at size 2, through time 4 |
| `single-event-config.json` | `single-event` | one red event at the origin, through time 3 |

```bash
./run run pyramid --new
./run run pyramid-size2 --new
./run run single-event --new
```
