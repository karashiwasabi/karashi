// File: static/js/inout.js (Corrected)
import { initHeader } from './inout_header.js';
import { initDetailsTable, getDetailsData, clearDetailsTable, populateDetailsTable } from './inout_details_table.js';

/**
 * Initializes all In/Out screen functionality and correctly wires up dependencies.
 */
export async function initInOut() {
  // Initialize the details table module first.
  initDetailsTable();

  // Initialize the header module, passing it the functions it needs from the details module.
  await initHeader(getDetailsData, clearDetailsTable, populateDetailsTable);
}