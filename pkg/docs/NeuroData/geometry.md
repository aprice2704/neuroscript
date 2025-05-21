# NeuroData Geometry & Scene Formats Specification

:: type: Specification
:: version: 0.1.0
:: status: draft
:: dependsOn: docs/metadata.md, docs/references.md, docs/neurodata/symbolic_math.md, docs/neurodata/tree.md, docs/neurodata_and_composite_file_spec.md
:: howToUpdate: Refine attributes and syntax for each format, specify coordinate systems/units conventions, detail tool requirements.

## 1. Purpose

This document defines a set of NeuroData formats for representing 3D geometry and scene structures:
* **Polygon Meshes (`.ndmesh`):** Explicit representation using vertices and faces.
* **Constructive Solid Geometry (`.ndcsg`):** Implicit representation using boolean operations on primitive shapes.
* **Signed Distance Fields (`.ndsdf`):** Implicit representation using a distance function.
* **Scene Graphs (`.ndscenegraph`):** Hierarchical structure for organizing transformations, geometry, lights, and cameras.

These formats aim to provide structured, human-readable (where practical) representations suitable for use within the NeuroScript ecosystem, primarily intended for manipulation and rendering by dedicated tools.

## 2. Common Elements

* **Metadata:** Each format uses standard `:: key: value` metadata [cite: uploaded:neuroscript/docs/metadata.md], including `:: type: <TypeName>`, `:: version:`, and potentially `:: id:`, `:: description:`.
* **Units & Coordinate System:** A convention should be established (e.g., default to millimeters, right-handed Y-up coordinate system) or specified via metadata (e.g., `:: units: mm`, `:: coords: RHS_Y_UP`). This is crucial for interoperability.
* **References:** Standard NeuroScript references (`[ref:<location>[#<block_id>]]` [cite: generated previously in `docs/references.md`]) are used extensively, especially in Scene Graphs, to link nodes to geometry, materials, lights, cameras, etc.
* **Data Representation:** Large datasets (like mesh vertices) might be stored line-by-line or within fenced blocks. Functional representations (`.ndcsg`, `.ndsdf`) use fenced blocks. Scene graphs use indentation and tagged attribute lines.

## 3. Format Specifications

### 3.1 Polygon Mesh (`.ndmesh`)

* **Purpose:** Represents geometry as a collection of vertices and polygons (faces).
* `:: type: Mesh`
* **Structure:** Typically uses tagged lines or sections for vertices and faces.
* **Attributes/Sections:**
    * `VERTEX <x> <y> <z>`: Defines a vertex. Coordinates are floats.
    * `FACE <idx1> <idx2> <idx3> [<idx4>...]`: Defines a polygon face using 1-based indices into the vertex list. Supports triangles, quads, or n-gons (tool support may vary).
    * `NORMAL <nx> <ny> <nz>`: (Optional) Defines a vertex normal. Usually listed in the same order as vertices.
    * `UV <u v>`: (Optional) Defines a texture coordinate. Usually listed in the same order as vertices.
    * *(Alternative: Vertex, Face, Normal, UV data could be placed within fenced blocks, e.g., ```csv ... ``` or ```json ... ```, for large meshes).*
* **Example:**
    ```ndmesh
    :: type: Mesh
    :: version: 0.1.0
    :: id: simple-cube-mesh

    VERTEX 0 0 0; VERTEX 1 0 0; VERTEX 1 1 0; VERTEX 0 1 0
    VERTEX 0 0 1; VERTEX 1 0 1; VERTEX 1 1 1; VERTEX 0 1 1

    FACE 1 2 3 4; FACE 5 6 7 8; FACE 1 2 6 5
    FACE 2 3 7 6; FACE 3 4 8 7; FACE 4 1 5 8
    ```

### 3.2 Constructive Solid Geometry (`.ndcsg`)

* **Purpose:** Represents shapes by combining primitive solids using boolean operations.
* `:: type: CSG`
* **Structure:** A tree of operations and primitives, represented using Functional Notation (similar to `.ndmath`) within a fenced block.
* **Syntax:** Defines functions for primitives, operations, and transformations.
    * **Primitives:** `Sphere(radius)`, `Cube(size | Vector(sx, sy, sz))`, `Cylinder(radius, height)`...
    * **Operations:** `Union(obj1, obj2, ...)`, `Difference(obj1, obj2, ...)`, `Intersection(obj1, obj2, ...)`
    * **Transformations:** `Translate(Vector(tx, ty, tz), obj)`, `Rotate(Vector(ax, ay, az), angle_degrees, obj)`, `Scale(Vector(sx, sy, sz), obj)`
    * **Helper:** `Vector(x, y, z)`
* **Example:**
    ```ndcsg
    :: type: CSG
    :: version: 0.1.0
    :: id: csg-example

    ```funcgeom
    Difference(
      Cube(size=20), # Assuming centered cube if Vector not used
      Sphere(radius=13)
    )
    ```
    ```

### 3.3 Signed Distance Field (`.ndsdf`)

* **Purpose:** Represents shapes implicitly via a function `f(x, y, z)` yielding the shortest signed distance to the surface.
* `:: type: SDF`
* **Structure:** The core is a mathematical expression defining the distance function. Uses the `.ndmath` Functional Notation.
* **Syntax:** Contains a fenced block storing the `.ndmath` expression.
* **Example:**
    ```ndsdf
    :: type: SDF
    :: version: 0.1.0
    :: id: sphere-sdf-example

    ```funcmath
    # Defines f(x, y, z) = length(Vector(x,y,z)) - radius
    Subtract( Length(Vector(x, y, z)), 5.0 )
    ```
    ```

### 3.4 Scene Graph (`.ndscenegraph`)

* **Purpose:** Represents a hierarchical structure of nodes containing transformations, references to content (geometry, lights, cameras), materials, and potentially grouping other nodes.
* `:: type: SceneGraph`
* **Structure:** Uses an indentation-based tree structure (like `.ndtree`). Each node is defined by `NODE <node_id>` followed by indented tagged attribute lines (like `.ndform`).
* **Node Attributes:**
    * `LABEL "<friendly_name>"`: (Optional) Human-readable name.
    * `TRANSFORM <transform_definition>`: Node's transformation relative to parent. Uses Functional Notation (e.g., `Translate(...) Rotate(...) Scale(...)`).
    * `CONTENT_REF "[ref:<location>#<content_id>]"`: Reference to `.ndmesh`, `.ndcsg`, `.ndsdf`, `.ndlight`, `.ndcamera`, etc.
    * `MATERIAL_REF "[ref:<material_id>]"`: (Optional) Reference to a material definition (presumed `.ndmat` format).
    * `PROPERTY <key> <value>`: (Optional) Custom user data.
* **Example:**
    ```ndscenegraph
    :: type: SceneGraph
    :: version: 0.1.0
    :: id: simple-scene

    NODE root
      NODE object1
        LABEL "Translated Cube"
        TRANSFORM Translate(5, 0, 0)
        CONTENT_REF "[ref:this#simple-cube-mesh]" # Refers to mesh defined elsewhere in same composite doc
      NODE object2
        LABEL "Rotated Sphere Group"
        TRANSFORM Rotate(0, 1, 0, 45)
        NODE sphere_visual # Child node
          LABEL "Actual Sphere"
          CONTENT_REF "[ref:geometry/primitives.ndsdf#unit-sphere]" # Refers to SDF in another file
          MATERIAL_REF "[ref:materials.ndmat#blue_plastic]"
      NODE main_camera
        TRANSFORM Translate(0, 10, 20) LookAt(0, 0, 0)
        CONTENT_REF "[ref:cameras.ndcamera#default]"
    ```

## 5. Tooling Requirements

Using these geometry and scene formats requires a significant suite of **new** NeuroScript tools, likely wrapping external geometry processing, CAS, and/or rendering libraries:
* **Parsing Tools:** `TOOL.ParseMesh`, `TOOL.ParseCSG`, `TOOL.ParseSDF` (using `TOOL.MathFromFunctional`), `TOOL.ParseSceneGraph`.
* **Evaluation/Processing Tools:**
    * `TOOL.EvaluateCSG` -> `.ndmesh` (Convert CSG to mesh).
    * `TOOL.MeshFromSDF` -> `.ndmesh` (e.g., Marching Cubes).
    * `TOOL.EvaluateSDF(sdf_ref, x, y, z)` -> `distance`.
    * `TOOL.ProcessSceneGraph` (e.g., flatten transforms, collect renderable objects).
    * `TOOL.MeshBoolean(mesh_a, mesh_b, operation)`
    * `TOOL.SimplifyMesh`, `TOOL.ValidateMesh`, `TOOL.CalculateMeshVolume`, etc.
* **Rendering Tools:** `TOOL.RenderSceneGraph(scene_ref, camera_ref)` -> `image_data` (Could be complex, potentially invoking external renderers).
* **Conversion Tools:** `TOOL.ExportMesh(mesh_ref, format)` (e.g., format="stl", "obj").
