export function getApi() {
  return window.go?.api?.API;
}

export async function listRoots() {
  return await getApi().ListRoots();
}

export async function listDirectory(path) {
  return await getApi().ListDirectory(path);
}

export async function addCopyJob(sources, dest, conflict) {
  return await getApi().AddCopyJob(sources, dest, conflict);
}

export async function addMoveJob(sources, dest, conflict) {
  return await getApi().AddMoveJob(sources, dest, conflict);
}

export async function addDeleteJob(sources) {
  return await getApi().AddDeleteJob(sources);
}

export async function addCompressJob(sources, dest, conflict) {
  return await getApi().AddCompressJob(sources, dest, conflict);
}

export async function addExtractJob(archive, dest, conflict) {
  return await getApi().AddExtractJob(archive, dest, conflict);
}

export async function getJobs() {
  return await getApi().GetJobs();
}

export async function cancelJob(jobId) {
  return await getApi().CancelJob(jobId);
}
