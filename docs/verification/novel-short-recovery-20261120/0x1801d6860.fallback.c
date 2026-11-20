/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801d6860 */

longlong FUN_1801d6860(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  undefined8 *puVar2;
  longlong lVar3;
  int *piVar4;
  char cVar5;
  int iVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong lVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong *plVar12;
  code *pcVar13;
  code *pcVar14;
  undefined8 uVar15;
  longlong *plVar16;
  longlong *local_res18;
  longlong *local_res20;
  longlong *local_60;
  longlong *local_50;
  longlong *local_48;

  plVar11 = DAT_1803bc830;
  plVar12 = (longlong *)0x0;
  plVar1 = (longlong *)*param_3;
  plVar16 = (longlong *)0x0;
  local_48 = (longlong *)0x0;
  local_60 = (longlong *)0x0;
  local_res18 = DAT_180391dd8;
  local_res20 = DAT_1803bc830;
  if (DAT_1803bc830 == (longlong *)0x0) {
LAB_1801d68dc:
    DAT_1803bc830 = (longlong *)FUN_18036cd10(param_1,DAT_1803bc8f0,DAT_1803ca5e8,0x20);
  }
  else if ((1 < *DAT_1803bc830) || (DAT_1803bc830[2] != 0)) {
    if ((DAT_1803bc830 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bc830 && (*DAT_1803bc830 = *DAT_1803bc830 + -1, *plVar11 == 0)))) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    goto LAB_1801d68dc;
  }
  plVar11 = DAT_1803bc830;
  local_res20 = DAT_1803bc830;
  FUN_180001d00(param_1,DAT_1803bc830);
  uVar15 = DAT_1803bcb20;
  plVar7 = (longlong *)FUN_18037f0c0();
  if (plVar7 == (longlong *)0x0) {
LAB_1801d693b:
    plVar10 = *(longlong **)(param_1 + 0x60);
    uVar15 = 0xdc;
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar7 = plVar12;
    local_60 = plVar12;
  }
  else {
    plVar8 = (longlong *)FUN_18000d280();
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (plVar8 == (longlong *)0x0) goto LAB_1801d693b;
    local_50 = (longlong *)FUN_18036d750(plVar8);
    plVar7 = (longlong *)
             FUN_180367890(FUN_1801d78e0,DAT_1803ca5e8,DAT_1803bcb70,DAT_1803bcb78,DAT_1803bc900,
                           &local_50,1,0x448);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    local_50 = (longlong *)PyUnicode_Join(uVar15,plVar7);
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    plVar7 = local_50;
    piVar4 = DAT_1803bcac8;
    if (local_50 == (longlong *)0x0) goto LAB_1801d693b;
    plVar8 = (longlong *)FUN_180372160(param_1,0x11);
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    *(int **)plVar8[3] = piVar4;
    piVar4 = DAT_1803bcb30;
    lVar9 = *(longlong *)(param_1 + 0x10);
    plVar11 = *(longlong **)(lVar9 + 0x41da8);
    if (plVar11 == (longlong *)0x0) {
      plVar11 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
    }
    else {
      *(longlong *)(lVar9 + 0x41da8) = plVar11[3];
      *(int *)(lVar9 + 0x41e40) = *(int *)(lVar9 + 0x41e40) + -1;
      *plVar11 = 1;
    }
    plVar10 = plVar11 + -2;
    plVar11[3] = 0;
    plVar11[4] = 0;
    plVar11[5] = 0;
    lVar9 = _PyThreadState_GetCurrent();
    lVar9 = *(longlong *)(*(longlong *)(lVar9 + 0x10) + 0xd0);
    puVar2 = *(undefined8 **)(lVar9 + 8);
    *puVar2 = plVar10;
    plVar11[-1] = plVar11[-1] & 3;
    plVar11[-1] = plVar11[-1] | (ulonglong)puVar2;
    *plVar10 = lVar9;
    *(longlong **)(lVar9 + 8) = plVar10;
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    plVar11[3] = (longlong)piVar4;
    iVar6 = (int)*local_50 + 1;
    if (iVar6 != 0) {
      *(int *)plVar7 = iVar6;
    }
    plVar11[4] = (longlong)plVar7;
    piVar4 = DAT_1803bcb38;
    if (*DAT_1803bcb38 + 1 != 0) {
      *DAT_1803bcb38 = *DAT_1803bcb38 + 1;
    }
    plVar11[5] = (longlong)piVar4;
    lVar9 = PyUnicode_Join(DAT_1803bc980,plVar11);
    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    if (lVar9 == 0) {
      uVar15 = 0xdf;
      goto LAB_1801d6eb3;
    }
    *(longlong *)(plVar8[3] + 8) = lVar9;
    piVar4 = DAT_1803bcb40;
    lVar9 = *(longlong *)(param_1 + 0x10);
    plVar11 = *(longlong **)(lVar9 + 0x41db8);
    if (plVar11 == (longlong *)0x0) {
      plVar11 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,5);
    }
    else {
      *(longlong *)(lVar9 + 0x41db8) = plVar11[3];
      *(int *)(lVar9 + 0x41e48) = *(int *)(lVar9 + 0x41e48) + -1;
      *plVar11 = 1;
    }
    plVar10 = plVar11 + -2;
    plVar11[3] = 0;
    plVar11[4] = 0;
    plVar11[5] = 0;
    plVar11[6] = 0;
    plVar11[7] = 0;
    lVar9 = _PyThreadState_GetCurrent();
    lVar9 = *(longlong *)(lVar9 + 0x10);
    lVar3 = *(longlong *)(lVar9 + 0xd0);
    puVar2 = *(undefined8 **)(lVar3 + 8);
    *puVar2 = plVar10;
    plVar11[-1] = plVar11[-1] & 3;
    plVar11[-1] = plVar11[-1] | (ulonglong)puVar2;
    *plVar10 = lVar3;
    *(longlong **)(lVar3 + 8) = plVar10;
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    plVar11[3] = (longlong)piVar4;
    plVar10 = (longlong *)FUN_18037f0c0(lVar9,plVar1,DAT_1803bc998);
    if (plVar10 == (longlong *)0x0) {
LAB_1801d6c8b:
      plVar10 = *(longlong **)(param_1 + 0x60);
      uVar15 = 0xe0;
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
LAB_1801d6ebb:
      plVar11 = local_res20;
      local_60 = plVar12;
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, local_60 = plVar16, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
        plVar11 = local_res20;
      }
    }
    else {
      lVar9 = FUN_180375520(param_1,plVar10,DAT_1803bc980);
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      if (lVar9 == 0) goto LAB_1801d6c8b;
      plVar11[4] = lVar9;
      piVar4 = DAT_1803bca40;
      if (*DAT_1803bca40 + 1 != 0) {
        *DAT_1803bca40 = *DAT_1803bca40 + 1;
      }
      plVar11[5] = (longlong)piVar4;
      plVar10 = (longlong *)FUN_18037f0c0(piVar4,plVar1,DAT_1803bc9a8);
      if (plVar10 == (longlong *)0x0) goto LAB_1801d6c8b;
      lVar9 = FUN_180375520(param_1,plVar10,DAT_1803bc980);
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      if (lVar9 == 0) goto LAB_1801d6c8b;
      plVar11[6] = lVar9;
      piVar4 = DAT_1803bcb48;
      if (*DAT_1803bcb48 + 1 != 0) {
        *DAT_1803bcb48 = *DAT_1803bcb48 + 1;
      }
      plVar11[7] = (longlong)piVar4;
      lVar9 = PyUnicode_Join(DAT_1803bc980,plVar11);
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      if (lVar9 == 0) {
        uVar15 = 0xe0;
        goto LAB_1801d6eb3;
      }
      *(longlong *)(plVar8[3] + 0x10) = lVar9;
      piVar4 = DAT_1803bcb50;
      if (*DAT_1803bcb50 + 1 != 0) {
        *DAT_1803bcb50 = *DAT_1803bcb50 + 1;
      }
      *(int **)(plVar8[3] + 0x18) = piVar4;
      piVar4 = DAT_1803bc980;
      if (*DAT_1803bc980 + 1 != 0) {
        *DAT_1803bc980 = *DAT_1803bc980 + 1;
      }
      *(int **)(plVar8[3] + 0x20) = piVar4;
      lVar9 = FUN_1801d3800(param_1);
      if (lVar9 == 0) {
        FUN_180373340(param_1,&local_res18,DAT_1803bcae0);
        uVar15 = 0xe3;
        plVar10 = local_res18;
        goto LAB_1801d6ebb;
      }
      uVar15 = 0xe3;
      *(undefined4 *)(local_res20 + 5) = 0xe3;
      lVar9 = FUN_180376210(param_1,lVar9);
      if (lVar9 == 0) {
LAB_1801d6eb3:
        plVar10 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1801d6ebb;
      }
      *(longlong *)(plVar8[3] + 0x28) = lVar9;
      piVar4 = DAT_1803bc980;
      if (*DAT_1803bc980 + 1 != 0) {
        *DAT_1803bc980 = *DAT_1803bc980 + 1;
      }
      *(int **)(plVar8[3] + 0x30) = piVar4;
      piVar4 = DAT_1803bc970;
      if (*DAT_1803bc970 + 1 != 0) {
        *DAT_1803bc970 = *DAT_1803bc970 + 1;
      }
      *(int **)(plVar8[3] + 0x38) = piVar4;
      lVar9 = FUN_18037f0c0(piVar4,plVar1,DAT_1803bc978);
      if (lVar9 == 0) {
        uVar15 = 0xe6;
        goto LAB_1801d6eb3;
      }
      *(longlong *)(plVar8[3] + 0x40) = lVar9;
      piVar4 = DAT_1803bc980;
      if (*DAT_1803bc980 + 1 != 0) {
        *DAT_1803bc980 = *DAT_1803bc980 + 1;
      }
      *(int **)(plVar8[3] + 0x48) = piVar4;
      piVar4 = DAT_1803bcae8;
      if (*DAT_1803bcae8 + 1 != 0) {
        *DAT_1803bcae8 = *DAT_1803bcae8 + 1;
      }
      *(int **)(plVar8[3] + 0x50) = piVar4;
      lVar9 = FUN_18037f0c0(piVar4,plVar1,DAT_1803bcaf0);
      if (lVar9 == 0) {
        uVar15 = 0xe9;
        goto LAB_1801d6eb3;
      }
      *(longlong *)(plVar8[3] + 0x58) = lVar9;
      piVar4 = DAT_1803bc980;
      if (*DAT_1803bc980 + 1 != 0) {
        *DAT_1803bc980 = *DAT_1803bc980 + 1;
      }
      *(int **)(plVar8[3] + 0x60) = piVar4;
      piVar4 = DAT_1803bcb58;
      if (*DAT_1803bcb58 + 1 != 0) {
        *DAT_1803bcb58 = *DAT_1803bcb58 + 1;
      }
      *(int **)(plVar8[3] + 0x68) = piVar4;
      lVar9 = FUN_18037f0c0(piVar4,plVar1,DAT_1803bcb60);
      if (lVar9 == 0) {
        uVar15 = 0xec;
        goto LAB_1801d6eb3;
      }
      *(longlong *)(plVar8[3] + 0x70) = lVar9;
      piVar4 = DAT_1803bc980;
      if (*DAT_1803bc980 + 1 != 0) {
        *DAT_1803bc980 = *DAT_1803bc980 + 1;
      }
      *(int **)(plVar8[3] + 0x78) = piVar4;
      piVar4 = DAT_1803bca68;
      if (*DAT_1803bca68 + 1 != 0) {
        *DAT_1803bca68 = *DAT_1803bca68 + 1;
      }
      *(int **)(plVar8[3] + 0x80) = piVar4;
      local_48 = plVar8;
      plVar11 = (longlong *)FUN_18037f0c0(piVar4,plVar1,DAT_1803bcb28);
      if (plVar11 == (longlong *)0x0) {
        plVar10 = *(longlong **)(param_1 + 0x60);
        uVar15 = 0xf0;
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar11 = local_res20;
        local_60 = plVar16;
      }
      else {
        plVar16 = (longlong *)FUN_18000d280();
        local_res18 = plVar16;
        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        if (plVar16 == (longlong *)0x0) {
          plVar10 = *(longlong **)(param_1 + 0x60);
          uVar15 = 0xf0;
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar11 = local_res20;
          local_60 = plVar12;
        }
        else {
          do {
            pcVar13 = *(code **)(plVar16[1] + 0xe0);
            if (pcVar13 == (code *)0x0) {
              PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                           *(undefined8 *)(plVar16[1] + 0x18));
LAB_1801d7555:
              plVar10 = *(longlong **)(param_1 + 0x60);
              if (plVar10 != (longlong *)0x0) {
                plVar11 = (longlong *)plVar10[1];
                if ((*(uint *)((longlong *)plVar10[1] + 0x15) & 0x40000000) == 0) {
                  plVar11 = plVar10;
                }
                if ((plVar11 != *(longlong **)PyExc_StopIteration_exref) &&
                   (((-1 < *(int *)(plVar11[1] + 0xa8) ||
                     ((*(uint *)(plVar11 + 0x15) & 0x40000000) == 0)) ||
                    (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  uVar15 = 0xf0;
                  if (local_60 != (longlong *)0x0) goto LAB_1801d7414;
                  goto LAB_1801d7432;
                }
                *(undefined8 *)(param_1 + 0x60) = 0;
                if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                }
              }
              if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
                 (*local_60 = *local_60 + -1, *local_60 == 0)) {
                (**(code **)(local_60[1] + 0x30))(local_60);
              }
              if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                (**(code **)(plVar16[1] + 0x30))(plVar16);
              }
              lVar9 = PyUnicode_Join(DAT_1803bc9e0,plVar8);
              if (lVar9 != 0) {
                FUN_180001d80(param_1);
                if ((-1 < (int)*local_50) && (*local_50 = *local_50 + -1, *local_50 == 0)) {
                  (**(code **)(plVar7[1] + 0x30))();
                }
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
                   (*local_60 = *local_60 + -1, *local_60 == 0)) {
                  (**(code **)(local_60[1] + 0x30))(local_60);
                }
                if ((int)*plVar1 < 0) {
                  return lVar9;
                }
                *plVar1 = *plVar1 + -1;
                if (*plVar1 != 0) {
                  return lVar9;
                }
                (**(code **)(plVar1[1] + 0x30))();
                return lVar9;
              }
              plVar10 = *(longlong **)(param_1 + 0x60);
              uVar15 = 0xf7;
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar11 = local_res20;
              goto LAB_1801d7654;
            }
            plVar11 = (longlong *)(*pcVar13)(plVar16);
            if (plVar11 == (longlong *)0x0) goto LAB_1801d7555;
            if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
               (*local_60 = *local_60 + -1, *local_60 == 0)) {
              (**(code **)(local_60[1] + 0x30))(local_60);
            }
            iVar6 = (int)*plVar11 + 1;
            if (iVar6 != 0) {
              *(int *)plVar11 = iVar6;
            }
            if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
               (*local_60 = *local_60 + -1, *local_60 == 0)) {
              (**(code **)(local_60[1] + 0x30))();
            }
            piVar4 = DAT_1803bca98;
            lVar9 = *(longlong *)(param_1 + 0x10);
            plVar12 = *(longlong **)(lVar9 + 0x41da8);
            if (plVar12 == (longlong *)0x0) {
              plVar12 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
            }
            else {
              *(longlong *)(lVar9 + 0x41da8) = plVar12[3];
              *(int *)(lVar9 + 0x41e40) = *(int *)(lVar9 + 0x41e40) + -1;
              *plVar12 = 1;
            }
            plVar10 = plVar12 + -2;
            plVar12[3] = 0;
            plVar12[4] = 0;
            plVar12[5] = 0;
            lVar9 = _PyThreadState_GetCurrent();
            lVar9 = *(longlong *)(*(longlong *)(lVar9 + 0x10) + 0xd0);
            puVar2 = *(undefined8 **)(lVar9 + 8);
            *puVar2 = plVar10;
            plVar12[-1] = plVar12[-1] & 3;
            plVar12[-1] = plVar12[-1] | (ulonglong)puVar2;
            *plVar10 = lVar9;
            *(longlong **)(lVar9 + 8) = plVar10;
            if (*piVar4 + 1 != 0) {
              *piVar4 = *piVar4 + 1;
            }
            plVar12[3] = (longlong)piVar4;
            lVar9 = FUN_180375520(param_1,plVar11,DAT_1803bc980);
            local_60 = plVar11;
            if (lVar9 == 0) {
              plVar10 = *(longlong **)(param_1 + 0x60);
              uVar15 = 0xf2;
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
              }
              goto LAB_1801d7414;
            }
            plVar12[4] = lVar9;
            piVar4 = DAT_1803bcaa0;
            if (*DAT_1803bcaa0 + 1 != 0) {
              *DAT_1803bcaa0 = *DAT_1803bcaa0 + 1;
            }
            plVar12[5] = (longlong)piVar4;
            lVar9 = PyUnicode_Join(DAT_1803bc980,plVar12);
            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
              (**(code **)(plVar12[1] + 0x30))(plVar12);
            }
            if (lVar9 == 0) {
              plVar10 = *(longlong **)(param_1 + 0x60);
              uVar15 = 0xf2;
              *(undefined8 *)(param_1 + 0x60) = 0;
              goto LAB_1801d7414;
            }
            plVar12 = (longlong *)FUN_180372160(param_1,4);
            *(longlong *)plVar12[3] = lVar9;
            piVar4 = DAT_1803bcaa8;
            if (*DAT_1803bcaa8 + 1 != 0) {
              *DAT_1803bcaa8 = *DAT_1803bcaa8 + 1;
            }
            *(int **)(plVar12[3] + 8) = piVar4;
            piVar4 = DAT_1803bca98;
            lVar9 = *(longlong *)(param_1 + 0x10);
            plVar16 = *(longlong **)(lVar9 + 0x41da8);
            if (plVar16 == (longlong *)0x0) {
              plVar16 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
            }
            else {
              *(longlong *)(lVar9 + 0x41da8) = plVar16[3];
              *(int *)(lVar9 + 0x41e40) = *(int *)(lVar9 + 0x41e40) + -1;
              *plVar16 = 1;
            }
            plVar10 = plVar16 + -2;
            plVar16[3] = 0;
            plVar16[4] = 0;
            plVar16[5] = 0;
            lVar9 = _PyThreadState_GetCurrent();
            lVar9 = *(longlong *)(*(longlong *)(lVar9 + 0x10) + 0xd0);
            puVar2 = *(undefined8 **)(lVar9 + 8);
            *puVar2 = plVar10;
            plVar16[-1] = plVar16[-1] & 3;
            plVar16[-1] = plVar16[-1] | (ulonglong)puVar2;
            *plVar10 = lVar9;
            *(longlong **)(lVar9 + 8) = plVar10;
            if (*piVar4 + 1 != 0) {
              *piVar4 = *piVar4 + 1;
            }
            plVar16[3] = (longlong)piVar4;
            lVar9 = FUN_180375520(param_1,plVar11,DAT_1803bc980);
            if (lVar9 == 0) {
              uVar15 = 0xf4;
LAB_1801d749d:
              plVar10 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                (**(code **)(plVar16[1] + 0x30))(plVar16);
              }
LAB_1801d74c2:
              plVar16 = local_res18;
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
                plVar16 = local_res18;
              }
              goto LAB_1801d7414;
            }
            plVar16[4] = lVar9;
            piVar4 = DAT_1803bcab0;
            if (*DAT_1803bcab0 + 1 != 0) {
              *DAT_1803bcab0 = *DAT_1803bcab0 + 1;
            }
            plVar16[5] = (longlong)piVar4;
            lVar9 = PyUnicode_Join(DAT_1803bc980,plVar16);
            if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
              (**(code **)(plVar16[1] + 0x30))(plVar16);
            }
            if (lVar9 == 0) {
              plVar10 = *(longlong **)(param_1 + 0x60);
              uVar15 = 0xf4;
              *(undefined8 *)(param_1 + 0x60) = 0;
              goto LAB_1801d74c2;
            }
            *(longlong *)(plVar12[3] + 0x10) = lVar9;
            piVar4 = DAT_1803bcab8;
            lVar9 = *(longlong *)(param_1 + 0x10);
            plVar16 = *(longlong **)(lVar9 + 0x41da8);
            if (plVar16 == (longlong *)0x0) {
              plVar16 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
            }
            else {
              *(longlong *)(lVar9 + 0x41da8) = plVar16[3];
              *(int *)(lVar9 + 0x41e40) = *(int *)(lVar9 + 0x41e40) + -1;
              *plVar16 = 1;
            }
            plVar10 = plVar16 + -2;
            plVar16[3] = 0;
            plVar16[4] = 0;
            plVar16[5] = 0;
            lVar9 = _PyThreadState_GetCurrent();
            lVar9 = *(longlong *)(*(longlong *)(lVar9 + 0x10) + 0xd0);
            puVar2 = *(undefined8 **)(lVar9 + 8);
            *puVar2 = plVar10;
            plVar16[-1] = plVar16[-1] & 3;
            plVar16[-1] = plVar16[-1] | (ulonglong)puVar2;
            *plVar10 = lVar9;
            *(longlong **)(lVar9 + 8) = plVar10;
            if (*piVar4 + 1 != 0) {
              *piVar4 = *piVar4 + 1;
            }
            plVar16[3] = (longlong)piVar4;
            lVar9 = FUN_180375520(param_1,plVar11,DAT_1803bc980);
            if (lVar9 == 0) {
              uVar15 = 0xf5;
              goto LAB_1801d749d;
            }
            plVar16[4] = lVar9;
            piVar4 = DAT_1803bcac0;
            if (*DAT_1803bcac0 + 1 != 0) {
              *DAT_1803bcac0 = *DAT_1803bcac0 + 1;
            }
            plVar16[5] = (longlong)piVar4;
            lVar9 = PyUnicode_Join(DAT_1803bc980,plVar16);
            if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
              (**(code **)(plVar16[1] + 0x30))(plVar16);
            }
            if (lVar9 == 0) {
              plVar10 = *(longlong **)(param_1 + 0x60);
              uVar15 = 0xf5;
              *(undefined8 *)(param_1 + 0x60) = 0;
              goto LAB_1801d74c2;
            }
            *(longlong *)(plVar12[3] + 0x18) = lVar9;
            cVar5 = FUN_180372360(param_1,plVar8,plVar12);
            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
              (**(code **)(plVar12[1] + 0x30))(plVar12);
            }
            if (cVar5 == '\0') {
              plVar10 = *(longlong **)(param_1 + 0x60);
              uVar15 = 0xf1;
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar16 = local_res18;
              goto LAB_1801d7414;
            }
            lVar9 = *(longlong *)(param_1 + 0x10);
            if ((((*(int *)(*(longlong *)(lVar9 + 0x60) + 0x4d4) != 0) ||
                 (*(int *)(lVar9 + 0x1a0) != 0)) && (iVar6 = Py_MakePendingCalls(), iVar6 < 0)) &&
               (*(longlong *)(param_1 + 0x60) != 0)) goto LAB_1801d73fd;
            if (*(int *)(lVar9 + 0x174) != 0) {
              PyEval_SaveThread();
              PyEval_AcquireThread(param_1);
            }
            plVar16 = local_res18;
          } while (*(longlong *)(param_1 + 0x80) == 0);
          plVar11 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x80) = 0;
          uVar15 = FUN_180373490(param_1);
          *(undefined8 *)(param_1 + 0x60) = uVar15;
          if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
             (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
LAB_1801d73fd:
          plVar10 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          uVar15 = 0xf0;
          plVar16 = local_res18;
LAB_1801d7414:
          if ((-1 < (int)*local_60) && (*local_60 = *local_60 + -1, *local_60 == 0)) {
            (**(code **)(local_60[1] + 0x30))(local_60);
          }
LAB_1801d7432:
          plVar11 = local_res20;
          if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
            (**(code **)(plVar16[1] + 0x30))(plVar16);
            plVar11 = local_res20;
          }
        }
      }
    }
  }
LAB_1801d7654:
  piVar4 = (int *)plVar10[5];
  if (piVar4 == (int *)0x0) {
    pcVar13 = (code *)FUN_18038e2a0(plVar11,uVar15);
    plVar16 = (longlong *)plVar10[5];
    pcVar14 = pcVar13;
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
LAB_1801d76ae:
      pcVar13 = pcVar14;
      if ((pcVar14 != (code *)0x0) && (*(int *)pcVar14 + 1 != 0)) {
        *(int *)pcVar14 = *(int *)pcVar14 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar4 + 6) == plVar11) goto LAB_1801d76f7;
    pcVar13 = (code *)FUN_18038e2a0(plVar11,uVar15);
    *(int **)(pcVar13 + 0x10) = piVar4;
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    plVar16 = (longlong *)plVar10[5];
    pcVar14 = pcVar13;
    if (pcVar13 != _Py_NoneStruct_exref) goto LAB_1801d76ae;
    pcVar14 = (code *)0x0;
  }
  plVar10[5] = (longlong)pcVar14;
  if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))();
  }
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
LAB_1801d76f7:
  plVar12 = local_48;
  FUN_18036d0a0(plVar11,&DAT_180391c1c,plVar1,plVar7,local_48,local_60);
  plVar16 = DAT_1803bc830;
  if (plVar11 == DAT_1803bc830) {
    if ((-1 < (int)*DAT_1803bc830) && (*DAT_1803bc830 = *DAT_1803bc830 + -1, *plVar16 == 0)) {
      (**(code **)(plVar16[1] + 0x30))();
    }
    DAT_1803bc830 = (longlong *)0x0;
  }
  lVar9 = **(longlong **)(param_1 + 0x38);
  plVar11 = *(longlong **)(lVar9 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar9 + 8);
  *(undefined4 *)(plVar11 + 8) = 0xffffffff;
  plVar16 = (longlong *)plVar11[2];
  if (((plVar16 != (longlong *)0x0) && (plVar11[2] = 0, -1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))();
  }
  if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))(plVar11);
  }
  plVar11[10] = 0;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))(plVar12);
  }
  if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
     (*local_60 = *local_60 + -1, *local_60 == 0)) {
    (**(code **)(local_60[1] + 0x30))(local_60);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar10;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
