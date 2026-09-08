// ── Theme ──────────────────────────────────────────────────────────────────
const themeSelect = document.getElementById('theme-select');

function applyTheme(name) {
	document.documentElement.setAttribute('data-theme', name);
	localStorage.setItem('cap-theme', name);
	themeSelect.value = name;
}

themeSelect.addEventListener('change', function() { applyTheme(themeSelect.value); });
applyTheme(localStorage.getItem('cap-theme') || 'cap');

// ── Kind chips ─────────────────────────────────────────────────────────────
const KIND_COLORS = {
	context:       '#4a7fd4',
	capability:    '#3daa6e',
	concept:       '#d4a535',
	invariant:     '#e05c5c',
	scenario:      '#9a6dd4',
	specification: '#3db8c8',
	verification:  '#7aad4a',
	adr:           '#c07840',
	task:          '#888888',
};
const STATUS_COLORS = {
	done:          '#3daa6e',
	'in-progress': '#5b8af5',
	proposed:      '#d4a535',
	draft:         '#7880a8',
};

const activeKinds = new Set(Object.keys(KIND_COLORS));

function updateKindState() {
	document.getElementById('kind-state').value = [...activeKinds].join(',');
}
updateKindState();

const chipsEl = document.getElementById('chips');
Object.entries(KIND_COLORS).forEach(function(entry) {
	const k = entry[0];
	const c = entry[1];
	const chip = document.createElement('span');
	chip.className = 'kind-chip';
	chip.dataset.kind = k;
	chip.style.cssText = 'background:' + c + '20;border-color:' + c + ';color:' + c;
	chip.innerHTML = '<span class="dot" style="background:' + c + '"></span>' + k;
	chip.addEventListener('click', function() {
		const allKinds = Object.keys(KIND_COLORS);
		const isSolo = activeKinds.size === 1 && activeKinds.has(k);
		if (isSolo) {
			allKinds.forEach(function(kind) { activeKinds.add(kind); });
		} else {
			allKinds.forEach(function(kind) { activeKinds.delete(kind); });
			activeKinds.add(k);
		}
		document.querySelectorAll('.kind-chip').forEach(function(el) {
			el.classList.toggle('off', !activeKinds.has(el.dataset.kind));
		});
		updateKindState();
		htmx.trigger(document.body, 'kindchange');
		updateGraphSel();
	});
	chipsEl.appendChild(chip);
});

// ── Graph (D3) ─────────────────────────────────────────────────────────────
let gZoom, gLinks, gNodes, gSim, gNodeMap, gSelId;
let gNeighbours = {};

function nodeRadius(d) {
	if (d.kind === 'context')    return 20;
	if (d.kind === 'capability') return 14;
	if (d.kind === 'scenario')   return 13;
	if (d.kind === 'concept')    return 10;
	return 8;
}

async function initGraph() {
	const data = await (await fetch('/graph')).json();
	const svgEl = document.getElementById('graph-svg');
	const svg   = d3.select(svgEl);
	const W = svgEl.clientWidth, H = svgEl.clientHeight;

	svg.append('defs').append('marker')
		.attr('id', 'arr').attr('viewBox', '0 -4 8 8').attr('refX', 8).attr('refY', 0)
		.attr('markerWidth', 5).attr('markerHeight', 5).attr('orient', 'auto')
		.append('path').attr('d', 'M0,-4L8,0L0,4').attr('fill', 'context-stroke');

	gZoom = d3.zoom().scaleExtent([.08, 4]).on('zoom', function(e) { g.attr('transform', e.transform); });
	svg.call(gZoom);
	svg.on('click', function() {
		gSelId = null;
		updateGraphSel();
		document.getElementById('detail-panel').classList.remove('open');
	});

	const g = svg.append('g');
	const nodes = data.nodes.map(function(n) { return Object.assign({}, n); });
	const links = data.edges.map(function(e) { return {source: e.source, target: e.target}; });

	// Build neighbour index so updateGraphSel can keep connected nodes visible.
	data.nodes.forEach(function(n) { gNeighbours[n.id] = new Set(); });
	data.edges.forEach(function(e) {
		(gNeighbours[e.source] = gNeighbours[e.source] || new Set()).add(e.target);
		(gNeighbours[e.target] = gNeighbours[e.target] || new Set()).add(e.source);
	});

	gSim = d3.forceSimulation(nodes)
		.force('link', d3.forceLink(links).id(function(d) { return d.id; }).distance(72).strength(.45))
		.force('charge', d3.forceManyBody().strength(-230))
		.force('center', d3.forceCenter(W / 2, H / 2))
		.force('col', d3.forceCollide().radius(function(d) { return nodeRadius(d) + 8; }));

	gLinks = g.append('g').selectAll('line')
		.data(links).join('line').attr('class', 'glink');

	gNodes = g.append('g').selectAll('g')
		.data(nodes).join('g').attr('class', 'gnode')
		.call(d3.drag()
			.on('start', function(e, d) { if (!e.active) gSim.alphaTarget(.3).restart(); d.fx = d.x; d.fy = d.y; })
			.on('drag',  function(e, d) { d.fx = e.x; d.fy = e.y; })
			.on('end',   function(e, d) { if (!e.active) gSim.alphaTarget(0); d.fx = null; d.fy = null; })
		);

	gNodes.append('circle')
		.attr('r', nodeRadius)
		.attr('fill', function(d) { return KIND_COLORS[d.kind] || '#888'; })
		.attr('stroke', function(d) { return d3.color(KIND_COLORS[d.kind] || '#888').darker(.5); })
		.on('click', function(e, d) {
			e.stopPropagation();
			gSelId = d.id;
			updateGraphSel();
			document.getElementById('detail-panel').classList.add('open');
			htmx.ajax('GET', '/entities/' + encodeURIComponent(d.id), {target: '#detail-panel', swap: 'innerHTML'});
		});

	gNodes.append('text')
		.attr('dy', function(d) { return nodeRadius(d) + 9; })
		.attr('text-anchor', 'middle')
		.attr('fill', 'var(--muted)')
		.attr('font-size', function(d) { return d.kind === 'context' ? '10px' : '8px'; })
		.attr('pointer-events', 'none')
		.text(function(d) { return d.title.length > 16 ? d.title.slice(0, 16) + '…' : d.title; });

	gSim.on('tick', function() {
		gLinks
			.attr('x1', function(d) { return d.source.x; })
			.attr('y1', function(d) { return d.source.y; })
			.attr('x2', function(d) {
				const dx = d.target.x - d.source.x, dy = d.target.y - d.source.y;
				const len = Math.sqrt(dx * dx + dy * dy) || 1;
				return d.target.x - dx / len * nodeRadius(d.target);
			})
			.attr('y2', function(d) {
				const dx = d.target.x - d.source.x, dy = d.target.y - d.source.y;
				const len = Math.sqrt(dx * dx + dy * dy) || 1;
				return d.target.y - dy / len * nodeRadius(d.target);
			});
		gNodes.attr('transform', function(d) { return 'translate(' + d.x + ',' + d.y + ')'; });
	});

	gNodeMap = Object.fromEntries(nodes.map(function(n) { return [n.id, n]; }));

	document.getElementById('btn-fit').onclick = fitGraph;
	document.getElementById('btn-rst').onclick = function() {
		d3.select(svgEl).transition().duration(400).call(gZoom.transform, d3.zoomIdentity);
	};
}

function updateGraphSel() {
	if (!gNodes) return;
	const nb = gSelId ? (gNeighbours[gSelId] || new Set()) : null;
	gNodes.classed('sel', function(d) { return d.id === gSelId; });
	if (nb != null) {
		// A node is selected: dim everything outside its neighbourhood,
		// ignoring the kind filter so the full context is always visible.
		gNodes.classed('dim', function(d) { return d.id !== gSelId && !nb.has(d.id); });
		gLinks.classed('hi',  function(d) { return d.source.id === gSelId || d.target.id === gSelId; });
		gLinks.classed('dim', function(d) { return d.source.id !== gSelId && d.target.id !== gSelId; });
	} else {
		// No selection: dim nodes whose kind is not active.
		gNodes.classed('dim', function(d) { return !activeKinds.has(d.kind); });
		gLinks.classed('hi',  false);
		gLinks.classed('dim', function(d) { return !activeKinds.has(d.source.kind) || !activeKinds.has(d.target.kind); });
	}
}

function panToNode(id) {
	if (!gNodeMap || !gNodeMap[id]) return;
	const n = gNodeMap[id];
	const svgEl = document.getElementById('graph-svg');
	const W = svgEl.clientWidth, H = svgEl.clientHeight;
	const k = Math.max(d3.zoomTransform(svgEl).k, 0.6);
	d3.select(svgEl).transition().duration(400)
		.call(gZoom.transform, d3.zoomIdentity
			.translate(W / 2 - k * n.x, H / 2 - k * n.y)
			.scale(k));
}

function fitGraph() {
	if (!gSim) return;
	const svgEl = document.getElementById('graph-svg');
	const W = svgEl.clientWidth, H = svgEl.clientHeight;
	const ns = gSim.nodes();
	const x0 = Math.min.apply(null, ns.map(function(n) { return n.x; }));
	const x1 = Math.max.apply(null, ns.map(function(n) { return n.x; }));
	const y0 = Math.min.apply(null, ns.map(function(n) { return n.y; }));
	const y1 = Math.max.apply(null, ns.map(function(n) { return n.y; }));
	const sc = Math.min(.88 * W / (x1 - x0 + 60), .88 * H / (y1 - y0 + 60), 2);
	d3.select(svgEl).transition().duration(500)
		.call(gZoom.transform, d3.zoomIdentity
			.translate(W / 2 - sc * (x0 + x1) / 2, H / 2 - sc * (y0 + y1) / 2)
			.scale(sc));
}

// ── Monaco ─────────────────────────────────────────────────────────────────
let monacoEditor = null;
require.config({ paths: { vs: '/static/vendor/vs' } });
require(['vs/editor/editor.main'], function() { window.monacoReady = true; });

document.body.addEventListener('htmx:afterSwap', function(e) {
	if (e.detail.target && e.detail.target.id === 'detail-panel') {
		document.getElementById('detail-panel').classList.add('open');
	}
	const wrap = document.getElementById('editor-wrap');
	if (!wrap) return;
	if (monacoEditor) { monacoEditor.dispose(); monacoEditor = null; }
	const content = wrap.dataset.content || '';
	if (window.monacoReady) {
		monacoEditor = monaco.editor.create(wrap, {
			value:                content,
			language:             'markdown',
			theme:                'vs-dark',
			minimap:              { enabled: false },
			fontSize:             12,
			lineNumbers:          'off',
			scrollBeyondLastLine: false,
			wordWrap:             'on',
			folding:              false,
		});
	}
});

async function saveEntity() {
	const wrap = document.getElementById('editor-wrap');
	if (!wrap || !monacoEditor) return;
	const id      = wrap.dataset.fileId;
	const content = monacoEditor.getValue();
	const resp    = await fetch('/entities/' + encodeURIComponent(id), {
		method:  'PUT',
		headers: { 'Content-Type': 'text/plain' },
		body:    content,
	});
	const btn = document.getElementById('btn-save');
	if (!btn) return;
	btn.textContent = resp.ok ? 'Saved' : 'Error';
	setTimeout(function() { btn.textContent = 'Save'; }, 1500);
}

// ── Tree expand/collapse ───────────────────────────────────────────────────
const expanded = new Set();

document.getElementById('tree-body').addEventListener('click', function(e) {
	const row = e.target.closest('.tr');
	if (!row) return;
	const id = row.dataset.id;
	if (e.target.closest('.tc-toggle') && row.dataset.hasKids) {
		if (expanded.has(id)) {
			expanded.delete(id);
		} else {
			expanded.add(id);
		}
		applyExpanded();
		e.stopPropagation();
		return;
	}
	gSelId = id;
	updateGraphSel();
	panToNode(id);
});

function applyExpanded() {
	document.querySelectorAll('.tr').forEach(function(row) {
		const pid = row.dataset.parent;
		if (!pid) return;
		const parentVisible = isAncestorChainExpanded(row);
		row.style.display = parentVisible ? '' : 'none';
	});
	document.querySelectorAll('.tr[data-has-kids]').forEach(function(row) {
		const toggle = row.querySelector('.tc-toggle');
		if (toggle) toggle.classList.toggle('open', expanded.has(row.dataset.id));
	});
}

function isAncestorChainExpanded(row) {
	const pid = row.dataset.parent;
	if (!pid) return true;
	if (!expanded.has(pid)) return false;
	const parentRow = document.querySelector('.tr[data-id="' + CSS.escape(pid) + '"]');
	if (!parentRow) return false;
	return isAncestorChainExpanded(parentRow);
}

document.body.addEventListener('htmx:afterSwap', function(e) {
	if (e.detail.target && e.detail.target.id === 'tree-body') {
		expanded.clear();
		document.querySelectorAll('.tr[data-depth="0"]').forEach(function(row) {
			expanded.add(row.dataset.id);
		});
		applyExpanded();
	}
});

// ── Boot ──────────────────────────────────────────────────────────────────
initGraph();
