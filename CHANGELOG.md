# 1.0.0 (2026-10-29)


### Bug Fixes

* **evcc:** add JSON tags to Summary for consistent camelCase API output ([c8d0bc6](https://github.com/mafeko/ladegast/commit/c8d0bc6bbe0d6c1689f87ad1959821792b7513ce))
* **web:** point getFriendStatus at the renamed /status endpoint, build into internal/webui/dist ([c996bf1](https://github.com/mafeko/ladegast/commit/c996bf1c2a878d6b35a5607d198a22b611db42e6))
* **web:** redirect / to /admin - root path had no route, rendered blank ([45afee1](https://github.com/mafeko/ladegast/commit/45afee1a121c2babd24d2b276584098c075d8b59))
* **web:** settlements failure no longer blocks the usage chart ([19134ae](https://github.com/mafeko/ladegast/commit/19134aedb049a28ee4ae17d4c7ca8b9a3f83a22b))


### Features

* **admin:** full friend CRUD, chosen PINs, copyable links and UX fixes ([f3d77b4](https://github.com/mafeko/ladegast/commit/f3d77b40665a60fc987e8d65349b607b0e79b806))
* **authtoken:** add signed session tokens for friend/admin cookies ([0f2e391](https://github.com/mafeko/ladegast/commit/0f2e391046195139c79227d312c0860f46f3ecec))
* **ci:** semantic versioning and release automation ([98dfd89](https://github.com/mafeko/ladegast/commit/98dfd8980411253f6f2771e339a1e1acfc217d71))
* **config:** load runtime configuration from environment variables ([e61a4d3](https://github.com/mafeko/ladegast/commit/e61a4d3bca27e180db5459a0d2c9baa1f2361f9c))
* **dev:** add mock EVCC server with YAML-driven demo data ([2546ad3](https://github.com/mafeko/ladegast/commit/2546ad305fdf1273ad49ffdce0d5db3b07e7f2a7))
* embed the frontend into the Go binary (internal/webui) ([da23c1a](https://github.com/mafeko/ladegast/commit/da23c1afac94be91606a3a5154583a54a3abda1e))
* **evcc:** add EVCC client and session aggregation ([e786691](https://github.com/mafeko/ladegast/commit/e7866918062322632d655335bce2d094ea31079b))
* **httpserver): add logout endpoints; test(httpserver:** cover auth edge cases ([ff0c1c3](https://github.com/mafeko/ladegast/commit/ff0c1c3d91954aac4a3e39e4c775deb37a41e996))
* **idgen:** add UUID and numeric PIN generation ([4bfd606](https://github.com/mafeko/ladegast/commit/4bfd606b6310dd426a5cdf862f48d576d9f9b2ad))
* per-friend settlements with chart markers and table totals ([da8015f](https://github.com/mafeko/ladegast/commit/da8015f4bdb215693066b5c8a0db3c44c49ba20e))
* **store:** add bbolt-backed persistence for friends, admins, rate-limit ([be9ffa9](https://github.com/mafeko/ladegast/commit/be9ffa9220375c2b8e40d49f008a5765ae6ef9b8))
* **store:** add DeleteFriend ([773a14c](https://github.com/mafeko/ladegast/commit/773a14c95c3b1cb99c3c7e82e97a263214a1d4f2))
* **web:** add friend and admin UI components, with tests for every component ([a35f41f](https://github.com/mafeko/ladegast/commit/a35f41fb45af3f2f2ceb7c0a1710513b39e131c0))
* **web:** add FriendPage/AdminPage/AdminViewAsPage views, wire up router ([1680de2](https://github.com/mafeko/ladegast/commit/1680de20cffcb2a77dcd775d21959e8bc5395f56))
* **web:** add typed API client and format helpers, with full unit tests ([17c1f6c](https://github.com/mafeko/ladegast/commit/17c1f6c62975abf082e986da1f7831792b74674e))
* **web:** mobile-first frontend, fix date picker dropdown on phones ([9f4abf0](https://github.com/mafeko/ladegast/commit/9f4abf0aae15196a97d30e274e684294db52e256))
* **web:** restyle the entire UI with Tailwind CSS ([e85eea1](https://github.com/mafeko/ladegast/commit/e85eea17326e27b9da33030f0465cf29603431fa))
* **web:** usage explorer with stacked bar chart and time range picker ([75f6d17](https://github.com/mafeko/ladegast/commit/75f6d171b2995765c679037c66a993b2de658daf))
* wire main entrypoint (config, store, EVCC client, HTTP server) ([c1565f8](https://github.com/mafeko/ladegast/commit/c1565f896e3bf673e37237557641146e5d8e6157))
