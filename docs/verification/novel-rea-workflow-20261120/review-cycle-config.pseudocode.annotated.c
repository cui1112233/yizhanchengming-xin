longlong * FUN_180228490(longlong param_1,undefined8 param_2,undefined8 *param_3)
{
  longlong *plVar1;
  int *piVar2;
  bool bVar3;
  longlong *plVar4;
  char cVar5;
  int iVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong lVar11;
  longlong *plVar12;
  longlong *plVar13;
  longlong *plVar14;
  code *pcVar15;
  undefined8 uVar16;
  longlong *plVar17;
  code *pcVar18;
  code *local_res18;
  longlong *local_res20;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_48;
  longlong *local_40;
  plVar4 = DAT_1803bed88;
  plVar10 = (longlong *)0x0;
  plVar17 = (longlong *)0x0;
  plVar1 = (longlong *)*param_3;
  pcVar15 = (code *)0x0;
  pcVar18 = (code *)0x0;
  local_78 = (longlong *)0x0;
  local_68 = (longlong *)0x0;
  local_70 = (longlong *)0x0;
  local_a0 = (longlong *)0x0;
  local_98 = (longlong *)0x0;
  local_res18 = (code *)0x0;
  local_a8 = (longlong *)0x0;
  local_res20 = (longlong *)0x0;
  if (DAT_1803bed88 == (longlong *)0x0) {
LAB_180228535:
    DAT_1803bed88 = (longlong *)FUN_18036cd10(param_1,DAT_1803bf4c0,DAT_1803ca5b0,0x40);
  }
  else if ((1 < *DAT_1803bed88) || (DAT_1803bed88[2] != 0)) {
    if ((DAT_1803bed88 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bed88 && (*DAT_1803bed88 = *DAT_1803bed88 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_180228535;
  }
  plVar4 = DAT_1803bed88;
  FUN_180001d00(param_1,DAT_1803bed88);
  uVar16 = 0x167;
  *(undefined4 *)(plVar4 + 5) = 0x167;
/* Analyst constant binding: novels.services.agent_settings#102="load" bytes=[4319480, 4319486] */
  plVar7 = (longlong *)FUN_18037e560(param_1,plVar1,DAT_1803bf1f0);
  if (plVar7 == (longlong *)0x0) {
    plVar12 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar17 = plVar10;
    local_res18 = pcVar15;
  }
  else {
    plVar8 = (longlong *)FUN_18037f0c0();
    uVar16 = 0x168;
    local_78 = plVar7;
    if (plVar8 == (longlong *)0x0) {
      plVar12 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar17 = plVar10;
      local_res18 = pcVar15;
    }
    else {
      *(undefined4 *)(plVar4 + 5) = 0x168;
      plVar9 = (longlong *)FUN_1803770d0(param_1,plVar8,DAT_1803bf0c0);
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      if (plVar9 == (longlong *)0x0) {
        plVar12 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar17 = plVar10;
        local_res18 = pcVar15;
      }
      else {
        iVar6 = PyObject_IsInstance(plVar9,PyDict_Type_exref);
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        if (iVar6 == -1) {
          plVar12 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          local_res18 = pcVar18;
        }
        else if (iVar6 == 0) {
          plVar10 = (longlong *)FUN_180370ec0(param_1);
LAB_1802286e3:
          lVar11 = FUN_180219ce0(param_1);
          local_68 = plVar10;
          if (lVar11 == 0) {
/* Analyst constant binding: novels.services.agent_settings#41="default_agent_settings" bytes=[4318442, 4318466] */
            FUN_180373340(param_1,&local_res20,DAT_1803bf008);
            uVar16 = 0x169;
            plVar12 = local_res20;
            local_res18 = pcVar18;
          }
          else {
            uVar16 = 0x169;
            *(undefined4 *)(plVar4 + 5) = 0x169;
            plVar8 = (longlong *)FUN_180376210(param_1,lVar11);
            if (plVar8 == (longlong *)0x0) {
              plVar12 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              local_res18 = pcVar18;
            }
            else {
/* Analyst constant binding: novels.services.agent_settings#13="reviewCycle" bytes=[4317971, 4317984] */
              plVar9 = (longlong *)FUN_180024e50(param_1,plVar8,DAT_1803bef28);
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
              }
              if (plVar9 == (longlong *)0x0) {
                plVar12 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                local_res18 = pcVar18;
              }
              else {
                plVar17 = (longlong *)FUN_180370ec0(param_1);
                local_res20 = plVar17;
                plVar8 = (longlong *)FUN_18037f0c0();
                uVar16 = 0x16b;
                local_70 = plVar9;
                if (plVar8 == (longlong *)0x0) {
                  plVar12 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  local_res18 = pcVar18;
                }
                else {
                  *(undefined4 *)(plVar4 + 5) = 0x16b;
                  plVar12 = (longlong *)FUN_180376210(param_1,plVar8);
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if (plVar12 == (longlong *)0x0) {
                    plVar12 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    local_res18 = pcVar18;
                  }
                  else {
                    plVar8 = (longlong *)FUN_180001930();
                    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                      (**(code **)(plVar12[1] + 0x30))(plVar12);
                    }
                    if (plVar8 == (longlong *)0x0) {
                      plVar12 = *(longlong **)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      local_res18 = pcVar18;
                    }
                    else {
                      do {
                        pcVar18 = *(code **)(plVar8[1] + 0xe0);
                        if (pcVar18 == (code *)0x0) {
                          PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                       "\'%s\' object is not an iterator",
                                       *(undefined8 *)(plVar8[1] + 0x18));
LAB_180229434:
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          if (plVar12 != (longlong *)0x0) {
                            plVar13 = (longlong *)plVar12[1];
                            if ((*(uint *)((longlong *)plVar12[1] + 0x15) & 0x40000000) == 0) {
                              plVar13 = plVar12;
                            }
                            if ((plVar13 != *(longlong **)PyExc_StopIteration_exref) &&
                               (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                 ((*(uint *)(plVar13 + 0x15) & 0x40000000) == 0)) ||
                                (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              if (local_a8 != (longlong *)0x0) goto LAB_1802294ac;
                              goto LAB_1802294d2;
                            }
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                              (**(code **)(plVar12[1] + 0x30))(plVar12);
                            }
                          }
                          if (((local_a8 != (longlong *)0x0) && (-1 < (int)*local_a8)) &&
                             (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
                            (**(code **)(local_a8[1] + 0x30))(local_a8);
                          }
                          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                            (**(code **)(plVar8[1] + 0x30))();
                          }
                          if (plVar17 != (longlong *)0x0) {
                            iVar6 = (int)*plVar17 + 1;
                            if (iVar6 != 0) {
                              *(int *)plVar17 = iVar6;
                            }
                            FUN_180001d80(param_1);
                            if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                              (**(code **)(plVar17[1] + 0x30))(plVar17);
                            }
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))();
                            }
                            if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                               (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                              (**(code **)(plVar10[1] + 0x30))();
                            }
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))();
                            }
                            if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0)) &&
                               (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                              (**(code **)(local_a0[1] + 0x30))(local_a0);
                            }
                            if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
                               (*local_98 = *local_98 + -1, *local_98 == 0)) {
                              (**(code **)(local_98[1] + 0x30))(local_98);
                            }
                            if (((local_res18 != (code *)0x0) && (-1 < *(int *)local_res18)) &&
                               (*(longlong *)local_res18 = *(longlong *)local_res18 + -1,
                               *(longlong *)local_res18 == 0)) {
                              (**(code **)(*(longlong *)(local_res18 + 8) + 0x30))();
                            }
                            if ((int)*plVar1 < 0) {
                              return plVar17;
                            }
                            *plVar1 = *plVar1 + -1;
                            if (*plVar1 != 0) {
                              return plVar17;
                            }
                            (**(code **)(plVar1[1] + 0x30))();
                            return plVar17;
                          }
                          plVar10 = (longlong *)
                                    PyUnicode_FromFormat
                                              ("cannot access local variable \'%U\' where it is not associated with a value"
/* Analyst constant binding: novels.services.agent_settings#107="result" bytes=[4319603, 4319611] */
                                               ,DAT_1803bf218);
                          plVar12 = (longlong *)
                                    FUN_180376770(param_1,*(undefined8 *)
                                                           PyExc_UnboundLocalError_exref,plVar10);
                          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                          }
                          FUN_18000d1d0(param_1,plVar12);
                          uVar16 = 0x175;
                          goto LAB_1802295c3;
                        }
                        plVar12 = (longlong *)(*pcVar18)();
                        if (plVar12 == (longlong *)0x0) goto LAB_180229434;
                        if (((local_a8 != (longlong *)0x0) && (-1 < (int)*local_a8)) &&
                           (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
                          (**(code **)(local_a8[1] + 0x30))(local_a8);
                        }
                        local_a8 = plVar12;
                        if (*(code **)(plVar12[1] + 0xd8) == (code *)0x0) {
                          iVar6 = PySequence_Check(plVar12);
                          if (iVar6 == 0) {
                            PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                         "cannot unpack non-iterable %s object",
                                         *(undefined8 *)(plVar12[1] + 0x18));
                          }
                          else {
                            plVar17 = (longlong *)FUN_180001220(PySeqIter_Type_exref);
                            plVar17[2] = 0;
                            iVar6 = (int)*plVar12 + 1;
                            if (iVar6 != 0) {
                              *(int *)plVar12 = iVar6;
                            }
                            plVar17[3] = (longlong)plVar12;
                            plVar12 = plVar17 + -2;
                            lVar11 = _PyThreadState_GetCurrent();
                            lVar11 = *(longlong *)(*(longlong *)(lVar11 + 0x10) + 0xd0);
                            plVar13 = *(longlong **)(lVar11 + 8);
                            *plVar13 = (longlong)plVar12;
                            plVar17[-1] = plVar17[-1] & 3;
                            plVar17[-1] = plVar17[-1] | (ulonglong)plVar13;
                            *plVar12 = lVar11;
                            *(longlong **)(lVar11 + 8) = plVar12;
                            if (plVar17 != (longlong *)0x0) goto LAB_180228990;
                          }
LAB_180229400:
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1802294ac;
                        }
                        plVar17 = (longlong *)(**(code **)(plVar12[1] + 0xd8))();
                        if (plVar17 == (longlong *)0x0) goto LAB_180229400;
                        lVar11 = *(longlong *)(plVar17[1] + 0xe0);
                        if ((lVar11 == 0) || (lVar11 == DAT_1803ca310)) {
                          PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                       "iter() returned non-iterator of type \'%s\'",
                                       *(undefined8 *)(plVar17[1] + 0x18));
                          if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                            (**(code **)(plVar17[1] + 0x30))(plVar17);
                          }
                          goto LAB_180229400;
                        }
LAB_180228990:
                        plVar13 = (longlong *)(**(code **)(plVar17[1] + 0xe0))(plVar17);
                        if (plVar13 == (longlong *)0x0) {
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          if ((plVar12 == (longlong *)0x0) || (lVar11 = plVar12[1], lVar11 == 0)) {
LAB_18022932e:
                            plVar10 = (longlong *)
                                      PyUnicode_FromFormat
                                                ("not enough values to unpack (expected %d, got %d)"
                                                 ,2,0);
                            plVar12 = (longlong *)
                                      FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,
                                                    plVar10);
                            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                              (**(code **)(plVar10[1] + 0x30))(plVar10);
                            }
                          }
                          else {
                            if ((*(uint *)(*(longlong *)(lVar11 + 8) + 0xa8) & 0x40000000) != 0) {
                              lVar11 = *(longlong *)(lVar11 + 8);
                            }
                            plVar10 = plVar12;
                            if ((lVar11 == *(longlong *)PyExc_StopIteration_exref) ||
                               (((*(int *)(*(longlong *)(lVar11 + 8) + 0xa8) < 0 &&
                                 ((*(uint *)(lVar11 + 0xa8) & 0x40000000) != 0)) &&
                                (cVar5 = FUN_180382640(), cVar5 == '\x01')))) {
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                              goto LAB_18022932e;
                            }
                            *(undefined8 *)(param_1 + 0x60) = 0;
                          }
                          plVar13 = (longlong *)0x0;
LAB_18022937a:
                          plVar14 = (longlong *)0x0;
LAB_18022937d:
                          if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                            (**(code **)(plVar17[1] + 0x30))(plVar17);
                          }
                          if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13)) &&
                             (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                            (**(code **)(plVar13[1] + 0x30))(plVar13);
                          }
                          if (((plVar14 != (longlong *)0x0) && (-1 < (int)*plVar14)) &&
                             (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                            (**(code **)(plVar14[1] + 0x30))(plVar14);
                          }
                          goto LAB_1802294ac;
                        }
                        plVar14 = (longlong *)(**(code **)(plVar17[1] + 0xe0))(plVar17);
                        if (plVar14 == (longlong *)0x0) {
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          if ((plVar12 == (longlong *)0x0) || (lVar11 = plVar12[1], lVar11 == 0)) {
LAB_180229251:
                            plVar10 = (longlong *)
                                      PyUnicode_FromFormat
                                                ("not enough values to unpack (expected %d, got %d)"
                                                 ,2,1);
                            plVar12 = (longlong *)
                                      FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,
                                                    plVar10);
                            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                              (**(code **)(plVar10[1] + 0x30))(plVar10);
                            }
                          }
                          else {
                            if ((*(uint *)(*(longlong *)(lVar11 + 8) + 0xa8) & 0x40000000) != 0) {
                              lVar11 = *(longlong *)(lVar11 + 8);
                            }
                            plVar10 = plVar12;
                            if ((lVar11 == *(longlong *)PyExc_StopIteration_exref) ||
                               (((*(int *)(*(longlong *)(lVar11 + 8) + 0xa8) < 0 &&
                                 ((*(uint *)(lVar11 + 0xa8) & 0x40000000) != 0)) &&
                                (cVar5 = FUN_180382640(), cVar5 == '\x01')))) {
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                              goto LAB_180229251;
                            }
                            *(undefined8 *)(param_1 + 0x60) = 0;
                          }
                          goto LAB_18022937a;
                        }
                        plVar12 = (longlong *)(**(code **)(plVar17[1] + 0xe0))(plVar17);
                        if (plVar12 != (longlong *)0x0) {
                          if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                            (**(code **)(plVar12[1] + 0x30))(plVar12);
                          }
                          plVar10 = (longlong *)
                                    PyUnicode_FromFormat
                                              ("too many values to unpack (expected %d)",2);
                          plVar12 = (longlong *)
                                    FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,
                                                  plVar10);
                          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                          }
                          goto LAB_18022937d;
                        }
                        plVar12 = *(longlong **)(param_1 + 0x60);
                        if ((plVar12 != (longlong *)0x0) && (lVar11 = plVar12[1], lVar11 != 0)) {
                          if ((*(uint *)(*(longlong *)(lVar11 + 8) + 0xa8) & 0x40000000) != 0) {
                            lVar11 = *(longlong *)(lVar11 + 8);
                          }
                          if (lVar11 == *(longlong *)PyExc_StopIteration_exref) {
                            *(undefined8 *)(param_1 + 0x60) = 0;
LAB_180228a5b:
                            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                              (**(code **)(plVar12[1] + 0x30))(plVar12);
                            }
                            goto LAB_180228a73;
                          }
                          if ((*(int *)(*(longlong *)(lVar11 + 8) + 0xa8) < 0) &&
                             ((*(uint *)(lVar11 + 0xa8) & 0x40000000) != 0)) {
                            cVar5 = FUN_180382640();
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if (cVar5 == '\x01') goto LAB_180228a5b;
                          }
                          else {
                            *(undefined8 *)(param_1 + 0x60) = 0;
                          }
                          goto LAB_18022937d;
                        }
LAB_180228a73:
                        if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                          (**(code **)(plVar17[1] + 0x30))(plVar17);
                        }
                        iVar6 = (int)*plVar13 + 1;
                        if (iVar6 != 0) {
                          *(int *)plVar13 = iVar6;
                        }
                        if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0)) &&
                           (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                          (**(code **)(local_a0[1] + 0x30))();
                        }
                        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                          (**(code **)(plVar13[1] + 0x30))(plVar13);
                        }
                        iVar6 = (int)*plVar14 + 1;
                        if (iVar6 != 0) {
                          *(int *)plVar14 = iVar6;
                        }
                        if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
                           (*local_98 = *local_98 + -1, *local_98 == 0)) {
                          (**(code **)(local_98[1] + 0x30))();
                        }
                        if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                          (**(code **)(plVar14[1] + 0x30))(plVar14);
                        }
                        local_a0 = plVar13;
                        local_98 = plVar14;
                        if (plVar10 == (longlong *)0x0) {
                          plVar17 = (longlong *)
                                    PyUnicode_FromFormat
                                              ("cannot access local variable \'%U\' where it is not associated with a value"
/* Analyst constant binding: novels.services.agent_settings#92="cycle" bytes=[4319344, 4319351] */
                                               ,DAT_1803bf1a0);
                          plVar12 = (longlong *)
                                    FUN_180376770(param_1,*(undefined8 *)
                                                           PyExc_UnboundLocalError_exref,plVar17);
                          if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                            (**(code **)(plVar17[1] + 0x30))(plVar17);
                          }
                          FUN_18000d1d0(param_1,plVar12);
                          uVar16 = 0x16c;
                          goto LAB_1802294ac;
                        }
                        plVar17 = (longlong *)FUN_18037f0c0();
                        if (plVar17 == (longlong *)0x0) {
LAB_1802290e0:
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          uVar16 = 0x16c;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1802294ac;
                        }
                        *(undefined4 *)(plVar4 + 5) = 0x16c;
                        local_48 = plVar13;
                        local_40 = plVar14;
                        pcVar18 = (code *)FUN_180377750(param_1,plVar17,&local_48);
                        if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                          (**(code **)(plVar17[1] + 0x30))(plVar17);
                        }
                        if (pcVar18 == (code *)0x0) goto LAB_1802290e0;
                        if (((local_res18 != (code *)0x0) && (-1 < *(int *)local_res18)) &&
                           (*(longlong *)local_res18 = *(longlong *)local_res18 + -1,
                           *(longlong *)local_res18 == 0)) {
                          (**(code **)(*(longlong *)(local_res18 + 8) + 0x30))();
                        }
                        iVar6 = PyObject_IsInstance(plVar14,PyBool_Type_exref);
                        local_res18 = pcVar18;
                        if (iVar6 == -1) {
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          uVar16 = 0x16d;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1802294ac;
                        }
                        if (iVar6 != 0) {
                          if (pcVar18 == _Py_TrueStruct_exref) {
LAB_180228bfe:
                            bVar3 = true;
LAB_180228c74:
                            plVar17 = local_res20;
                            pcVar18 = _Py_FalseStruct_exref;
                            if (bVar3) {
                              pcVar18 = _Py_TrueStruct_exref;
                            }
                            if (local_res20 == (longlong *)0x0) {
                              plVar17 = (longlong *)
                                        PyUnicode_FromFormat
                                                  (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
/* Analyst constant binding: novels.services.agent_settings#107="result" bytes=[4319603, 4319611] */
                                                  ,DAT_1803bf218,pcVar18);
                              plVar12 = (longlong *)
                                        FUN_180376770(param_1,*(undefined8 *)
                                                               PyExc_UnboundLocalError_exref,plVar17
                                                     );
                              if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0))
                              {
                                (**(code **)(plVar17[1] + 0x30))(plVar17);
                              }
                              FUN_18000d1d0(param_1,plVar12);
                              uVar16 = 0x16e;
                              goto LAB_1802294ac;
                            }
                            iVar6 = PyDict_SetItem(local_res20,plVar13);
                            if (iVar6 == 0) goto LAB_180228e21;
                          }
                          else {
                            if ((pcVar18 == _Py_FalseStruct_exref) ||
                               (pcVar18 == _Py_NoneStruct_exref)) {
LAB_180228c72:
                              bVar3 = false;
                              goto LAB_180228c74;
                            }
                            lVar11 = *(longlong *)(pcVar18 + 8);
                            if ((*(longlong *)(lVar11 + 0x60) == 0) ||
                               (pcVar15 = *(code **)(*(longlong *)(lVar11 + 0x60) + 0x48),
                               pcVar15 == (code *)0x0)) {
                              if ((*(undefined8 **)(lVar11 + 0x70) == (undefined8 *)0x0) ||
                                 (pcVar15 = (code *)**(undefined8 **)(lVar11 + 0x70),
                                 pcVar15 == (code *)0x0)) {
                                if ((*(undefined8 **)(lVar11 + 0x68) == (undefined8 *)0x0) ||
                                   (pcVar15 = (code *)**(undefined8 **)(lVar11 + 0x68),
                                   pcVar15 == (code *)0x0)) goto LAB_180228bfe;
                                lVar11 = (*pcVar15)(pcVar18);
                              }
                              else {
                                lVar11 = (*pcVar15)(pcVar18);
                              }
                            }
                            else {
                              iVar6 = (*pcVar15)(pcVar18);
                              lVar11 = (longlong)iVar6;
                            }
                            if (0 < lVar11) {
                              bVar3 = true;
                              goto LAB_180228c74;
                            }
                            if (lVar11 == 0) goto LAB_180228c72;
                          }
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          uVar16 = 0x16e;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1802294ac;
                        }
                        iVar6 = PyObject_IsInstance(plVar14,PyFloat_Type_exref);
                        if (iVar6 == -1) {
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          uVar16 = 0x16f;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1802294ac;
                        }
                        if (iVar6 != 0) {
                          if (*(code **)(pcVar18 + 8) == PyUnicode_Type_exref) {
                            plVar12 = (longlong *)PyFloat_FromString();
                          }
                          else {
                            plVar12 = (longlong *)PyNumber_Float(pcVar18);
                          }
                          plVar17 = local_res20;
                          if (plVar12 != (longlong *)0x0) {
                            if (local_res20 == (longlong *)0x0) {
                              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                              {
                                (**(code **)(plVar12[1] + 0x30))(plVar12);
                              }
                              plVar17 = (longlong *)
                                        PyUnicode_FromFormat
                                                  (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
/* Analyst constant binding: novels.services.agent_settings#107="result" bytes=[4319603, 4319611] */
                                                  ,DAT_1803bf218);
                              plVar12 = (longlong *)
                                        FUN_180376770(param_1,*(undefined8 *)
                                                               PyExc_UnboundLocalError_exref,plVar17
                                                     );
                              if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0))
                              {
                                (**(code **)(plVar17[1] + 0x30))(plVar17);
                              }
                              FUN_18000d1d0(param_1,plVar12);
                              uVar16 = 0x170;
                              goto LAB_1802294ac;
                            }
                            iVar6 = PyDict_SetItem(local_res20,plVar13,plVar12);
                            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                              (**(code **)(plVar12[1] + 0x30))(plVar12);
                            }
                            if (iVar6 == 0) goto LAB_180228e21;
                          }
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          uVar16 = 0x170;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1802294ac;
                        }
                        iVar6 = PyObject_IsInstance(plVar14,PyLong_Type_exref);
                        plVar17 = local_res20;
                        if (iVar6 == -1) {
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          uVar16 = 0x171;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1802294ac;
                        }
                        if (iVar6 != 0) {
                          plVar12 = (longlong *)PyNumber_Long(pcVar18);
                          plVar17 = local_res20;
                          if (plVar12 != (longlong *)0x0) {
                            if (local_res20 == (longlong *)0x0) {
                              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                              {
                                (**(code **)(plVar12[1] + 0x30))(plVar12);
                              }
                              plVar17 = (longlong *)
                                        PyUnicode_FromFormat
                                                  (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
/* Analyst constant binding: novels.services.agent_settings#107="result" bytes=[4319603, 4319611] */
                                                  ,DAT_1803bf218);
                              plVar12 = (longlong *)
                                        FUN_180376770(param_1,*(undefined8 *)
                                                               PyExc_UnboundLocalError_exref,plVar17
                                                     );
                              if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0))
                              {
                                (**(code **)(plVar17[1] + 0x30))(plVar17);
                              }
                              FUN_18000d1d0(param_1,plVar12);
                              uVar16 = 0x172;
                              goto LAB_1802294ac;
                            }
                            iVar6 = PyDict_SetItem(local_res20,plVar13,plVar12);
                            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                              (**(code **)(plVar12[1] + 0x30))(plVar12);
                            }
                            if (iVar6 == 0) goto LAB_180228e21;
                          }
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          uVar16 = 0x172;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1802294ac;
                        }
                        if (local_res20 == (longlong *)0x0) {
                          plVar17 = (longlong *)
                                    PyUnicode_FromFormat
                                              ("cannot access local variable \'%U\' where it is not associated with a value"
/* Analyst constant binding: novels.services.agent_settings#107="result" bytes=[4319603, 4319611] */
                                               ,DAT_1803bf218);
                          plVar12 = (longlong *)
                                    FUN_180376770(param_1,*(undefined8 *)
                                                           PyExc_UnboundLocalError_exref,plVar17);
                          if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                            (**(code **)(plVar17[1] + 0x30))(plVar17);
                          }
                          FUN_18000d1d0(param_1,plVar12);
                          uVar16 = 0x174;
                          goto LAB_1802294ac;
                        }
                        iVar6 = PyDict_SetItem(local_res20,plVar13,pcVar18);
                        if (iVar6 != 0) {
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          uVar16 = 0x174;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1802294ac;
                        }
LAB_180228e21:
                        lVar11 = *(longlong *)(param_1 + 0x10);
                        if ((((*(int *)(*(longlong *)(lVar11 + 0x60) + 0x4d4) != 0) ||
                             (*(int *)(lVar11 + 0x1a0) != 0)) &&
                            (iVar6 = Py_MakePendingCalls(), iVar6 < 0)) &&
                           (plVar12 = *(longlong **)(param_1 + 0x60), plVar12 != (longlong *)0x0))
                        goto LAB_18022901f;
                        if (*(int *)(lVar11 + 0x174) != 0) {
                          PyEval_SaveThread();
                          PyEval_AcquireThread(param_1);
                        }
                      } while (*(longlong *)(param_1 + 0x80) == 0);
                      plVar17 = *(longlong **)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x80) = 0;
                      plVar12 = (longlong *)FUN_180373490(param_1);
                      *(longlong **)(param_1 + 0x60) = plVar12;
                      if ((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) {
                        *plVar17 = *plVar17 + -1;
                        if (*plVar17 == 0) {
                          (**(code **)(plVar17[1] + 0x30))(plVar17);
                        }
                        plVar12 = *(longlong **)(param_1 + 0x60);
                      }
LAB_18022901f:
                      *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802294ac:
                      plVar17 = local_res20;
                      if ((-1 < (int)*local_a8) && (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
                        (**(code **)(local_a8[1] + 0x30))(local_a8);
                        plVar17 = local_res20;
                      }
LAB_1802294d2:
                      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                        (**(code **)(plVar8[1] + 0x30))();
                      }
                    }
                  }
                }
              }
            }
          }
        }
        else {
          plVar8 = (longlong *)FUN_18037f0c0();
          if (plVar8 == (longlong *)0x0) {
            plVar12 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            local_res18 = pcVar18;
          }
          else {
            *(undefined4 *)(plVar4 + 5) = 0x168;
            plVar10 = (longlong *)FUN_1803770d0(param_1,plVar8,DAT_1803bf0c0);
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            if (plVar10 != (longlong *)0x0) goto LAB_1802286e3;
            plVar12 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            local_res18 = pcVar18;
          }
        }
      }
    }
  }
LAB_1802295c3:
  piVar2 = (int *)plVar12[5];
  if (piVar2 == (int *)0x0) {
    pcVar18 = (code *)FUN_18038e2a0(plVar4,uVar16);
    plVar10 = (longlong *)plVar12[5];
    pcVar15 = pcVar18;
    if (pcVar18 == _Py_NoneStruct_exref) {
      pcVar15 = (code *)0x0;
    }
    else {
LAB_18022961d:
      pcVar18 = pcVar15;
      if ((pcVar15 != (code *)0x0) && (*(int *)pcVar15 + 1 != 0)) {
        *(int *)pcVar15 = *(int *)pcVar15 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar4) goto LAB_180229666;
    pcVar18 = (code *)FUN_18038e2a0(plVar4,uVar16);
    *(int **)(pcVar18 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar10 = (longlong *)plVar12[5];
    pcVar15 = pcVar18;
    if (pcVar18 != _Py_NoneStruct_exref) goto LAB_18022961d;
    pcVar15 = (code *)0x0;
  }
  plVar12[5] = (longlong)pcVar15;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  if (((pcVar18 != (code *)0x0) && (-1 < *(int *)pcVar18)) &&
     (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1, *(longlong *)pcVar18 == 0)) {
    (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
  }
LAB_180229666:
  FUN_18036d0a0(plVar4,"oooooooo",plVar1,plVar17,local_78,local_68,local_70,local_a0,local_98,
                local_res18);
  plVar10 = DAT_1803bed88;
  if (plVar4 == DAT_1803bed88) {
    if ((-1 < (int)*DAT_1803bed88) && (*DAT_1803bed88 = *DAT_1803bed88 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))();
    }
    DAT_1803bed88 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) &&
     (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
    (**(code **)(plVar17[1] + 0x30))(plVar17);
  }
  if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
     (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))();
  }
  if (((local_68 != (longlong *)0x0) && (-1 < (int)*local_68)) &&
     (*local_68 = *local_68 + -1, *local_68 == 0)) {
    (**(code **)(local_68[1] + 0x30))();
  }
  if (((local_70 != (longlong *)0x0) && (-1 < (int)*local_70)) &&
     (*local_70 = *local_70 + -1, *local_70 == 0)) {
    (**(code **)(local_70[1] + 0x30))();
  }
  if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0)) &&
     (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
    (**(code **)(local_a0[1] + 0x30))(local_a0);
  }
  if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
     (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))(local_98);
  }
  if (((local_res18 != (code *)0x0) && (-1 < *(int *)local_res18)) &&
     (*(longlong *)local_res18 = *(longlong *)local_res18 + -1, *(longlong *)local_res18 == 0)) {
    (**(code **)(*(longlong *)(local_res18 + 8) + 0x30))(local_res18);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar12;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
