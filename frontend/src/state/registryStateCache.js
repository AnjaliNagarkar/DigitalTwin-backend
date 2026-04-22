const registryStateByScope = {
  agriculture: createEmptyRegistryState(),
  population: createEmptyRegistryState(),
}

function createEmptyRegistryState() {
  return {
    records: [],
    loadedAt: 0,
    category: '',
    search: '',
    subFilters: {},
    sortKey: '',
    sortDir: 'asc',
    currentPage: 1,
  }
}

export function getRegistryState(scope = 'agriculture') {
  if (!registryStateByScope[scope]) {
    registryStateByScope[scope] = createEmptyRegistryState()
  }
  return registryStateByScope[scope]
}

export function resetRegistryState(scope = 'agriculture') {
  registryStateByScope[scope] = createEmptyRegistryState()
  return registryStateByScope[scope]
}
